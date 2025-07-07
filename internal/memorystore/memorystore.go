package memorystore

import (
	"context"
	"fmt"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/dbservice"
	apierror "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/rs/zerolog/log"
	"sync"
)

var mu sync.Mutex // guards memoryStore
var store []bool
var storeLen = 0
var initialized = false

func Init() {
	if initialized {
		panic("MemoryStore Init was called more than once")
	}

	// allocate the memory
	store = make([]bool, 1000000)
	storeLen = len(store)

	initialized = true
}

func GetCheckboxStatus(checkboxNbr int) (bool, error) {
	if !checkboxNbrValid(checkboxNbr) {
		log.Error().Msgf("invalid checkbox number for call GetCheckboxStatus(%d)", checkboxNbr)
		return false, apierror.InternalError(fmt.Sprintf("invalid checkbox number for call GetCheckboxStatus(%d", checkboxNbr))
	}

	// dont need to lock for reads
	checked := store[checkboxNbr]

	return checked, nil
}

func DoCheck(checkboxNbr int, checked bool) error {
	if !checkboxNbrValid(checkboxNbr) {
		log.Error().Msgf("invalid checkbox number for call DoCheck(%d, %t)", checkboxNbr, checked)
		return apierror.InternalError(fmt.Sprintf("invalid checkbox number for call DoCheck(%d, %t)", checkboxNbr, checked))
	}
	mu.Lock()
	store[checkboxNbr] = checked
	mu.Unlock()

	return nil
}

func LoadCheckboxesFromStore(ctx context.Context) apierror.APIError {
	newMemoryStore, err := dbservice.GetFullCheckboxStore(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get full checkbox store from database")
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, "failed to get full checkbox store from database")
	}

	mu.Lock()
	store = *newMemoryStore
	mu.Unlock()

	return nil
}

func checkboxNbrValid(checkboxNbr int) bool {
	return checkboxNbr >= 0 && checkboxNbr < storeLen
}
