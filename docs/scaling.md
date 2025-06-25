# SCALING CONSIDERATIONS

This is a list of other steps we could take, to potentially increase throughput or performance, 
at the various layers.
- uuid pre-generation pool server 
- db: separate storage pools for each partition
- db: separate storage pools for the hot table vs the rest, or also for the updates table 

BACKPRESSURE
- Need to have some way to detect a queue growing and not draining, or growing much faster than it's draining
- Need to shut down checkbox change requests if queue is growing endlessly

PARTIAL AVAILABILITY
- Existing clients can continue to function even if new clients are not allowed 
- Existing clients (and new clients) can display the read-only and even updated version of the world, even if changes are temp blocked

MAINTENANCE (to remove old data for performance)
- Set retention period of queue to no more than a day 
- Purge from UPDATE_T any entries older than a day 
- Purge from CLIENTS_T that havent interacted with the system in more than a day 
- Calculate metrics every N minutes and append new row to metrics table 


https://www.somethingsimilar.com/2013/01/14/notes-on-distributed-systems-for-young-bloods/
