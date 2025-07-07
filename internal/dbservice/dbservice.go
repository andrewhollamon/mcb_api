package dbservice

import (
	"context"
	"fmt"
	apierror "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"time"
)

// UpdateCheckbox updates the state of a checkbox identified by its number with the specified checked status.
// It returns an APIError if the operation fails, with contextual and stack trace information.
func UpdateCheckbox(ctx context.Context, checkboxNbr int, checked bool, userUuid uuid.UUID, requestUuid uuid.UUID) apierror.APIError {
	// Begin transaction
	tx, err := BeginTx(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("failed to begin transaction inside UpdateCheckbox(%d, %t, %v, %v)", checkboxNbr, checked, userUuid, requestUuid)
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, "failed to begin transaction")
	}

	// Ensure cleanup - rollback on error
	defer func() {
		if err != nil {
			rollbackerr := RollbackTx(ctx, tx)
			if rollbackerr != nil {
				log.Error().Err(rollbackerr).Msgf(
					"failed to rollback transaction inside UpdateCheckbox(%d, %t, %v, %v)", checkboxNbr, checked, userUuid, requestUuid,
				)
			}
		}
	}()

	// Update CHECKBOX_T table
	_, err = ExecTx(ctx, tx, "UPDATE MCB.CHECKBOX_T "+
		"SET CHECKED_STATE = $1 WHERE CHECKBOX_NBR = $2",
		checked, checkboxNbr)
	if err != nil {
		log.Error().Err(err).Msgf("failed to update checkbox_t inside UpdateCheckbox(%d, %t, %v, %v)", checkboxNbr, checked, userUuid, requestUuid)
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, "failed to update checkbox state")
	}

	// Update CHECKBOX_DETAILS_T table
	_, err = ExecTx(ctx, tx, "UPDATE MCB.CHECKBOX_DETAILS_T "+
		"SET LAST_UPDATED_BY = $1, LAST_REQUEST_ID = $2, LAST_UPDATED_DATE = $3 "+
		"WHERE CHECKBOX_NBR = $4", userUuid, requestUuid, time.Now(), checkboxNbr)
	if err != nil {
		log.Error().Err(err).Msgf("failed to update checkbox_details_t inside UpdateCheckbox(%d, %t, %v, %v)", checkboxNbr, checked, userUuid, requestUuid)
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, "failed to update checkbox details")
	}

	// Commit transaction
	err = CommitTx(ctx, tx)
	if err != nil {
		log.Error().Err(err).Msgf("failed to commit transaction inside UpdateCheckbox(%d, %t, %v, %v)", checkboxNbr, checked, userUuid, requestUuid)
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, "failed to commit transaction")
	}

	log.Debug().Msgf("UpdateCheckbox(%d, %t, %v, %v) completed successfully", checkboxNbr, checked, userUuid, requestUuid)
	return nil
}

func GetCheckboxStatus(ctx context.Context, checkboxNbr int) (bool, time.Time, apierror.APIError) {
	// Query both tables with a JOIN to get checkbox state and last updated date
	rows, err := Query(ctx,
		"SELECT c.CHECKED_STATE, d.LAST_UPDATED_DATE "+
			"FROM MCB.CHECKBOX_T c "+
			"JOIN MCB.CHECKBOX_DETAILS_T d ON c.CHECKBOX_NBR = d.CHECKBOX_NBR "+
			"WHERE c.CHECKBOX_NBR = $1",
		checkboxNbr)
	if err != nil {
		log.Error().Err(err).Msgf("failed to query checkbox status inside GetCheckboxStatus(%d)", checkboxNbr)
		return false, time.UnixMilli(0), apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, "failed to query checkbox status")
	}
	defer rows.Close()

	// Check if any rows were returned
	if !rows.Next() {
		log.Debug().Msgf("no checkbox found with number %d inside GetCheckboxStatus(%d)", checkboxNbr, checkboxNbr)
		return false, time.UnixMilli(0), apierror.NewAPIErrorFromCode(apierror.ErrRecordNotFound, "checkbox not found")
	}

	// Scan the result
	var checkedState bool
	var lastUpdatedDate time.Time
	err = rows.Scan(&checkedState, &lastUpdatedDate)
	if err != nil {
		log.Error().Err(err).Msgf("failed to scan checkbox status result inside GetCheckboxStatus(%d)", checkboxNbr)
		return false, time.UnixMilli(0), apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, "failed to scan checkbox status result")
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msgf("rows iteration error inside GetCheckboxStatus(%d)", checkboxNbr)
		return false, time.UnixMilli(0), apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, "database iteration error")
	}

	log.Debug().Msgf("GetCheckboxStatus(%d) completed successfully: checked=%t, lastUpdated=%v", checkboxNbr, checkedState, lastUpdatedDate)
	return checkedState, lastUpdatedDate, nil
}

func GetFullCheckboxStore(ctx context.Context) (*[]bool, apierror.APIError) {
	checkboxes := make([]bool, 1000000)

	rows, err := Query(ctx, "SELECT CHECKED_STATE FROM MCB.CHECKBOX_T ORDER BY CHECKBOX_NBR")
	if err != nil {
		log.Error().Err(err).Msgf("failed to query checkbox status inside GetFullCheckboxStore")
		return nil, apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, "failed to query checkbox status inside GetFullCheckboxStore")
	}
	defer rows.Close()

	// I would really prefer a database driver that let me control batch size and stream through the results, to reduce
	// memory pressure. However, pgx does not seem to offer that, so I'm just accepting the memory load. Real world
	// testing will determine how much of a problem this is.
	i := 0
	for rows.Next() {
		checked, err := rows.Values()
		if err != nil {
			log.Error().Err(err).Msgf("failed to read value from result inside GetFullCheckboxStore")
			return nil, apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, "failed to read value from result inside GetFullCheckboxStore")
		}
		checkboxes[i] = checked[0].(bool)
		i++
	}

	// if we didnt get exactly 1,000,000 rows, then something is badly wrong
	if i != 999999 {
		log.Error().Msgf("expected to get %d checkboxes, got %d", 999999, i)
		return nil, apierror.NewAPIErrorFromCode(apierror.ErrDatabaseError, fmt.Sprintf("expected to get %d checkboxes, got %d", 999999, i))
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msgf("rows iteration error inside GetFullCheckboxStore")
		return nil, apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, "rows iteration error inside GetFullCheckboxStore")
	}

	return &checkboxes, nil
}

func InitDbPool(ctx context.Context) apierror.APIError {
	err := InitializePool(ctx)
	if err != nil {
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, "failed to initialize the database pool")
	}
	return nil
}
