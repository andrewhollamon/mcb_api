# MILLION CHECKBOXES

This project is inspired by the One Million Checkboxes project as featured on Hacker News.

This is an excellent project to demonstrate a variety of architecture styles to support high volume and 
shared state.

This also gives an opportunity to explore different ways to handle contention and conflict in 
shared state.

## ARCHITECTURE GOALS VERSION 1

The goals of the architecture design are as follows:

1. Present a UI with a million checkboxes, that any user may click on, and set the state permanently to "checked".
2. Checkboxes cannot be "unchecked".
3. Requests to mark a checkbox as "checked" will be queued, and processed in order they appear in the queue.
4. Requests to mark a checkbox that has already had its state flipped to "checked" will have no effect on the shared state, and will return a FAILED response back to the UI.
5. Requests to mark a checkbox that succeed will return a SUCCESS response back to the UI.
6. The UI will have a panel that shows both requests and responses (ie, attempted check #28743, success #28743, etc), that is updated asynchronously.
7. Requests will be processed asynchronously by the back-end, and the back end will allow front-end to query whether the request succeeded or failed.
8. Hard limits will be put into place to ensure hosting/serving/billing costs do not exceed a fixed amount, and the UI will show when this happens.
9. Back end will be architected such that the different layers can be partitioned independently, to allow for effective scaling.
10. No authN or authZ used in this version.
11. Request Throttling will be implemented at the API layer.
12. Logging of request success, failure to a cheap simple append-only store.
13. Load-testing tools to make some attempt at seeing how much load system can take.



