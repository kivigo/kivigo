* [] add a method to check if the key exists in the database
* [] add a method to return the health of the database connection
* [] add a method to return the type of the database (e.g. redis, local, etc.)
* [] Add key builder
* [] Add key expiration (Add method to check if the current database supports expiration & add method to set expiration)
* [] Add database etcd

* [] Add TTL (Time-To-Live) / key expiration: Allow setting a lifetime for each key, with automatic deletion after expiration
* [] Add transactions / batch operations: Support atomic or grouped operations (set/get/delete multiple keys)
* [] Add events / hooks: Allow registering callbacks on operations (set, delete, expire, etc.)
* [] Add local cache with synchronization: Add in-memory cache with automatic invalidation/synchronization
* [] Add statistics and monitoring: Expose metrics on usage, latency, number of keys, etc.
* [] Add prefix or pattern search: Allow listing or searching keys by pattern or prefix
