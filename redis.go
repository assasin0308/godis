// godis
package godis

import (
	"errors"
	"time"
)

// Option connect options
type Option struct {
	// redis host
	Host string
	// redis port
	Port int
	// connect timeout
	ConnectionTimeout time.Duration
	// read timeout
	SoTimeout time.Duration
	// redis password,if empty,then without auth
	Password string
	// which db to connect
	Db int
}

// Redis redis client tool
type Redis struct {
	client      *client
	pipeline    *pipeline
	transaction *transaction
	dataSource  *Pool
}

// constructor for creating new redis
func NewRedis(option *Option) *Redis {
	client := newClient(option)
	return &Redis{client: client}
}

//Connect connect to redis
func (r *Redis) Connect() error {
	return r.client.connect()
}

//Close close redis connection
func (r *Redis) Close() error {
	if r.dataSource != nil {
		if r.client.broken {
			return r.dataSource.returnBrokenResourceObject(r)
		} else {
			return r.dataSource.returnResourceObject(r)
		}
	}
	if r != nil && r.client != nil {
		return r.client.close()
	}
	return nil
}

func (r *Redis) setDataSource(pool *Pool) {
	r.dataSource = pool
}

// Send send command to redis
func (r *Redis) Send(command protocolCommand, args ...[]byte) error {
	return r.client.sendCommand(command, args...)
}

// SendByStr send command to redis
func (r *Redis) SendByStr(command string, args ...[]byte) error {
	return r.client.sendCommandByStr(command, args...)
}

// Receive receive reply from redis
func (r *Redis) Receive() error {
	if r != nil && r.client != nil {
		return r.client.close()
	}
	return nil
}

func (r *Redis) checkIsInMultiOrPipeline() error {
	if r.client.isInMulti {
		return errors.New("cannot use Redis when in Multi. Please use Transaction or reset redis state")
	}
	if r.pipeline != nil && len(r.pipeline.pipelinedResponses) > 0 {
		return errors.New("cannot use Redis when in Pipeline. Please use Pipeline or reset redis state")
	}
	return nil
}

//<editor-fold desc="rediscommands">

// Set the string value as value of the key. The string can't be longer than 1073741824 bytes (1 //GB)
// return Status code reply
func (r *Redis) Set(key, value string) (string, error) {
	err := r.client.set(key, value)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

// Set the string value as value of the key. The string can't be longer than 1073741824 bytes (1 //GB).
// param nxxx NX|XX, NX -- Only set the key if it does not already exist. XX -- Only set the key if it already exist.
// param expx EX|PX, expire time units: EX = seconds; PX = milliseconds
// param time expire time in the units of <code>expx</code>
//return Status code reply
func (r *Redis) SetWithParamsAndTime(key, value, nxxx, expx string, time int64) (string, error) {
	err := r.client.setWithParamsAndTime(key, value, nxxx, expx, time)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//Get the value of the specified key. If the key does not exist null is returned. If the value
//stored at key is not a string an error is returned because GET can only handle string values.
//param key
//return Bulk reply
func (r *Redis) Get(key string) (string, error) {
	err := r.client.get(key)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Return the type of the value stored at key in form of a string. The type can be one of "none",
//"string", "list", "set". "none" is returned if the key does not exist. Time complexity: O(1)
//param key
//return Status code reply, specifically: "none" if the key does not exist "string" if the key
//        contains a String value "list" if the key contains a List value "set" if the key
//        contains a Set value "zset" if the key contains a Sorted Set value "hash" if the key
//        contains a Hash value
func (r *Redis) Type(key string) (string, error) {
	err := r.client.typeKey(key)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Set a timeout on the specified key. After the timeout the key will be automatically deleted by
//the server. A key with an associated timeout is said to be volatile in Redis terminology.
//
//Voltile keys are stored on disk like the other keys, the timeout is persistent too like all the
//other aspects of the dataset. Saving a dataset containing expires and stopping the server does
//not stop the flow of time as Redis stores on disk the time when the key will no longer be
//available as Unix time, and not the remaining seconds.
//
//Since Redis 2.1.3 you can update the value of the timeout of a key already having an expire
//set. It is also possible to undo the expire at all turning the key into a normal key using the
//{@link #persist(String) PERSIST} command.
//
//Time complexity: O(1)
//@see <a href="http://code.google.com/p/redis/wiki/ExpireCommand">ExpireCommand</a>
//param key
//param seconds
//return Integer reply, specifically: 1: the timeout was set. 0: the timeout was not set since
//        the key already has an associated timeout (this may happen only in Redis versions &lt;
//        2.1.3, Redis &gt;= 2.1.3 will happily update the timeout), or the key does not exist.
func (r *Redis) Expire(key string, seconds int) (int64, error) {
	err := r.client.expire(key, seconds)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//EXPIREAT works exctly like {@link #expire(String, int) EXPIRE} but instead to get the number of
//seconds representing the Time To Live of the key as a second argument (that is a relative way
//of specifying the TTL), it takes an absolute one in the form of a UNIX timestamp (Number of
//seconds elapsed since 1 Gen 1970).
//
//EXPIREAT was introduced in order to implement the Append Only File persistence mode so that
//EXPIRE commands are automatically translated into EXPIREAT commands for the append only file.
//Of course EXPIREAT can also used by programmers that need a way to simply specify that a given
//key should expire at a given time in the future.
//
//Since Redis 2.1.3 you can update the value of the timeout of a key already having an expire
//set. It is also possible to undo the expire at all turning the key into a normal key using the
//{@link #persist(String) PERSIST} command.
//
//Time complexity: O(1)
//@see <a href="http://code.google.com/p/redis/wiki/ExpireCommand">ExpireCommand</a>
//param key
//param unixTime
//return Integer reply, specifically: 1: the timeout was set. 0: the timeout was not set since
//        the key already has an associated timeout (this may happen only in Redis versions &lt;
//        2.1.3, Redis &gt;= 2.1.3 will happily update the timeout), or the key does not exist.
func (r *Redis) ExpireAt(key string, unixtime int64) (int64, error) {
	err := r.client.expireAt(key, unixtime)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//The TTL command returns the remaining time to live in seconds of a key that has an
//{@link #expire(String, int) EXPIRE} set. This introspection capability allows a Redis client to
//check how many seconds a given key will continue to be part of the dataset.
//param key
//return Integer reply, returns the remaining time to live in seconds of a key that has an
//        EXPIRE. In Redis 2.6 or older, if the Key does not exists or does not have an
//        associated expire, -1 is returned. In Redis 2.8 or newer, if the Key does not have an
//        associated expire, -1 is returned or if the Key does not exists, -2 is returned.
func (r *Redis) Ttl(key string) (int64, error) {
	err := r.client.ttl(key)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

// Pttl Like TTL this command returns the remaining time to live of a key that has an expire set,
// with the sole difference that TTL returns the amount of remaining time in seconds while PTTL returns it in milliseconds.
//In Redis 2.6 or older the command returns -1 if the key does not exist or if the key exist but has no associated expire.
//Starting with Redis 2.8 the return value in case of error changed:
//The command returns -2 if the key does not exist.
//The command returns -1 if the key exists but has no associated expire.
//
//Integer reply: TTL in milliseconds, or a negative value in order to signal an error (see the description above).
func (r *Redis) Pttl(key string) (int64, error) {
	err := r.client.pttl(key)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

// Setrange Overwrites part of the string stored at key, starting at the specified offset,
// for the entire length of value. If the offset is larger than the current length of the string at key,
// the string is padded with zero-bytes to make offset fit. Non-existing keys are considered as empty strings,
// so this command will make sure it holds a string large enough to be able to set value at offset.
// Note that the maximum offset that you can set is 229 -1 (536870911), as Redis Strings are limited to 512 megabytes.
// If you need to grow beyond this size, you can use multiple keys.
//
// Warning: When setting the last possible byte and the string value stored at key does not yet hold a string value,
// or holds a small string value, Redis needs to allocate all intermediate memory which can block the server for some time.
// On a 2010 MacBook Pro, setting byte number 536870911 (512MB allocation) takes ~300ms,
// setting byte number 134217728 (128MB allocation) takes ~80ms,
// setting bit number 33554432 (32MB allocation) takes ~30ms and setting bit number 8388608 (8MB allocation) takes ~8ms.
// Note that once this first allocation is done,
// subsequent calls to SETRANGE for the same key will not have the allocation overhead.
//
// Patterns
// Thanks to SETRANGE and the analogous GETRANGE commands, you can use Redis strings as a linear array with O(1) random access. This is a very fast and efficient storage in many real world use cases.
//
// Return value
// Integer reply: the length of the string after it was modified by the command.
func (r *Redis) Setrange(key string, offset int64, value string) (int64, error) {
	err := r.client.setrange(key, offset, value)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

// Getrange Warning: this command was renamed to GETRANGE, it is called SUBSTR in Redis versions <= 2.0.
// Returns the substring of the string value stored at key,
// determined by the offsets start and end (both are inclusive).
// Negative offsets can be used in order to provide an offset starting from the end of the string.
// So -1 means the last character, -2 the penultimate and so forth.
//
// The function handles out of range requests by limiting the resulting range to the actual length of the string.
//
// Return value
// Bulk string reply
func (r *Redis) Getrange(key string, startOffset, endOffset int64) (string, error) {
	err := r.client.getrange(key, startOffset, endOffset)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//GETSET is an atomic set this value and return the old value command. Set key to the string
//value and return the old value stored at key. The string can't be longer than 1073741824 bytes (1 GB).
//
//Time complexity: O(1)
//param key
//param value
//return Bulk reply
func (r *Redis) GetSet(key, value string) (string, error) {
	err := r.client.getSet(key, value)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//SETNX works exactly like {@link #set(String, String) SET} with the only difference that if the
//key already exists no operation is performed. SETNX actually means "SET if Not eXists".
//
//Time complexity: O(1)
//param key
//param value
//return Integer reply, specifically: 1 if the key was set 0 if the key was not set
func (r *Redis) Setnx(key, value string) (int64, error) {
	err := r.client.setnx(key, value)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//The command is exactly equivalent to the following group of commands:
//{@link #set(String, String) SET} + {@link #expire(String, int) EXPIRE}. The operation is
//atomic.
//
//Time complexity: O(1)
//param key
//param seconds
//param value
//return Status code reply
func (r *Redis) Setex(key string, seconds int, value string) (string, error) {
	err := r.client.setex(key, seconds, value)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//IDECRBY work just like {@link #decr(String) INCR} but instead to decrement by 1 the decrement
//is integer.
//
//INCR commands are limited to 64 bit signed integers.
//
//Note: this is actually a string operation, that is, in Redis there are not "integer" types.
//Simply the string stored at the key is parsed as a base 10 64 bit signed integer, incremented,
//and then converted back as a string.
//
//Time complexity: O(1)
//@see #incr(String)
//@see #decr(String)
//@see #incrBy(String, long)
//param key
//param integer
//return Integer reply, this commands will reply with the new value of key after the increment.
func (r *Redis) DecrBy(key string, decrement int64) (int64, error) {
	err := r.client.decrBy(key, decrement)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Decrement the number stored at key by one. If the key does not exist or contains a value of a
//wrong type, set the key to the value of "0" before to perform the decrement operation.
//
//INCR commands are limited to 64 bit signed integers.
//
//Note: this is actually a string operation, that is, in Redis there are not "integer" types.
//Simply the string stored at the key is parsed as a base 10 64 bit signed integer, incremented,
//and then converted back as a string.
//
//Time complexity: O(1)
//@see #incr(String)
//@see #incrBy(String, long)
//@see #decrBy(String, long)
//param key
//return Integer reply, this commands will reply with the new value of key after the increment.
func (r *Redis) Decr(key string) (int64, error) {
	err := r.client.decr(key)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//INCRBY work just like {@link #incr(String) INCR} but instead to increment by 1 the increment is
//integer.
//
//INCR commands are limited to 64 bit signed integers.
//
//Note: this is actually a string operation, that is, in Redis there are not "integer" types.
//Simply the string stored at the key is parsed as a base 10 64 bit signed integer, incremented,
//and then converted back as a string.
//
//Time complexity: O(1)
//@see #incr(String)
//@see #decr(String)
//@see #decrBy(String, long)
//param key
//param integer
//return Integer reply, this commands will reply with the new value of key after the increment.
func (r *Redis) IncrBy(key string, increment int64) (int64, error) {
	err := r.client.incrBy(key, increment)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//INCRBYFLOAT
//
//INCRBYFLOAT commands are limited to double precision floating point values.
//
//Note: this is actually a string operation, that is, in Redis there are not "double" types.
//Simply the string stored at the key is parsed as a base double precision floating point value,
//incremented, and then converted back as a string. There is no DECRYBYFLOAT but providing a
//negative value will work as expected.
//
//Time complexity: O(1)
//param key
//param value
//return Double reply, this commands will reply with the new value of key after the increment.
func (r *Redis) IncrByFloat(key string, increment float64) (float64, error) {
	err := r.client.incrByFloat(key, increment)
	if err != nil {
		return 0, err
	}
	return StringToFloat64Reply(r.client.getBulkReply())
}

//Increment the number stored at key by one. If the key does not exist or contains a value of a
//wrong type, set the key to the value of "0" before to perform the increment operation.
//
//INCR commands are limited to 64 bit signed integers.
//
//Note: this is actually a string operation, that is, in Redis there are not "integer" types.
//Simply the string stored at the key is parsed as a base 10 64 bit signed integer, incremented,
//and then converted back as a string.
//
//Time complexity: O(1)
//@see #incrBy(String, long)
//@see #decr(String)
//@see #decrBy(String, long)
//param key
//return Integer reply, this commands will reply with the new value of key after the increment.
func (r *Redis) Incr(key string) (int64, error) {
	err := r.client.incr(key)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//If the key already exists and is a string, this command appends the provided value at the end
//of the string. If the key does not exist it is created and set as an empty string, so APPEND
//will be very similar to SET in this special case.
//
//Time complexity: O(1). The amortized time complexity is O(1) assuming the appended value is
//small and the already present value is of any size, since the dynamic string library used by
//Redis will double the free space available on every reallocation.
//param key
//param value
//return Integer reply, specifically the total length of the string after the append operation.
func (r *Redis) Append(key, value string) (int64, error) {
	err := r.client.append(key, value)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return a subset of the string from offset start to offset end (both offsets are inclusive).
//Negative offsets can be used in order to provide an offset starting from the end of the string.
//So -1 means the last char, -2 the penultimate and so forth.
//
//The function handles out of range requests without raising an error, but just limiting the
//resulting range to the actual length of the string.
//
//Time complexity: O(start+n) (with start being the start index and n the total length of the
//requested range). Note that the lookup part of this command is O(1) so for small strings this
//is actually an O(1) command.
//param key
//param start
//param end
//return Bulk reply
func (r *Redis) Substr(key string, start, end int) (string, error) {
	err := r.client.substr(key, start, end)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Set the specified hash field to the specified value.
//
//If key does not exist, a new key holding a hash is created.
//
//<b>Time complexity:</b> O(1)
//param key
//param field
//param value
//return If the field already exists, and the HSET just produced an update of the value, 0 is
//        returned, otherwise if a new field is created 1 is returned.
func (r *Redis) Hset(key, field, value string) (int64, error) {
	err := r.client.hset(key, field, value)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//If key holds a hash, retrieve the value associated to the specified field.
//
//If the field is not found or the key does not exist, a special 'nil' value is returned.
//
//<b>Time complexity:</b> O(1)
//param key
//param field
//return Bulk reply
func (r *Redis) Hget(key, field string) (string, error) {
	err := r.client.hget(key, field)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Set the specified hash field to the specified value if the field not exists. <b>Time
//complexity:</b> O(1)
//param key
//param field
//param value
//return If the field already exists, 0 is returned, otherwise if a new field is created 1 is
//        returned.
func (r *Redis) Hsetnx(key, field, value string) (int64, error) {
	err := r.client.hsetnx(key, field, value)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Set the respective fields to the respective values. HMSET replaces old values with new values.
//
//If key does not exist, a new key holding a hash is created.
//
//<b>Time complexity:</b> O(N) (with N being the number of fields)
//param key
//param hash
//return Return OK or Exception if hash is empty
func (r *Redis) Hmset(key string, hash map[string]string) (string, error) {
	err := r.client.hmset(key, hash)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Retrieve the values associated to the specified fields.
//
//If some of the specified fields do not exist, nil values are returned. Non existing keys are
//considered like empty hashes.
//
//<b>Time complexity:</b> O(N) (with N being the number of fields)
//param key
//param fields
//return Multi Bulk Reply specifically a list of all the values associated with the specified
//        fields, in the same order of the request.
func (r *Redis) Hmget(key string, fields ...string) ([]string, error) {
	err := r.client.hmget(key, fields...)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Increment the number stored at field in the hash at key by value. If key does not exist, a new
//key holding a hash is created. If field does not exist or holds a string, the value is set to 0
//before applying the operation. Since the value argument is signed you can use this command to
//perform both increments and decrements.
//
//The range of values supported by HINCRBY is limited to 64 bit signed integers.
//
//<b>Time complexity:</b> O(1)
//param key
//param field
//param value
//return Integer reply The new value at field after the increment operation.
func (r *Redis) HincrBy(key, field string, value int64) (int64, error) {
	err := r.client.hincrBy(key, field, value)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Increment the number stored at field in the hash at key by a double precision floating point
//value. If key does not exist, a new key holding a hash is created. If field does not exist or
//holds a string, the value is set to 0 before applying the operation. Since the value argument
//is signed you can use this command to perform both increments and decrements.
//
//The range of values supported by HINCRBYFLOAT is limited to double precision floating point
//values.
//
//<b>Time complexity:</b> O(1)
//param key
//param field
//param value
//return Double precision floating point reply The new value at field after the increment
//        operation.
func (r *Redis) HincrByFloat(key, field string, value float64) (float64, error) {
	err := r.client.hincrByFloat(key, field, value)
	if err != nil {
		return 0, err
	}
	return StringToFloat64Reply(r.client.getBulkReply())
}

//Test for existence of a specified field in a hash. <b>Time complexity:</b> O(1)
//param key
//param field
//return Return 1 if the hash stored at key contains the specified field. Return 0 if the key is
//        not found or the field is not present.
func (r *Redis) Hexists(key, field string) (bool, error) {
	err := r.client.hexists(key, field)
	if err != nil {
		return false, err
	}
	return Int64ToBoolReply(r.client.getIntegerReply())
}

//Remove the specified field from an hash stored at key.
//
//<b>Time complexity:</b> O(1)
//param key
//param fields
//return If the field was present in the hash it is deleted and 1 is returned, otherwise 0 is
//        returned and no operation is performed.
func (r *Redis) Hdel(key string, fields ...string) (int64, error) {
	err := r.client.hdel(key, fields...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return the number of items in a hash.
//
//<b>Time complexity:</b> O(1)
//param key
//return The number of entries (fields) contained in the hash stored at key. If the specified
//        key does not exist, 0 is returned assuming an empty hash.
func (r *Redis) Hlen(key string) (int64, error) {
	err := r.client.hlen(key)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return all the fields in a hash.
//
//<b>Time complexity:</b> O(N), where N is the total number of entries
//param key
//return All the fields names contained into a hash.
func (r *Redis) Hkeys(key string) ([]string, error) {
	err := r.client.hkeys(key)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Return all the values in a hash.
//
//<b>Time complexity:</b> O(N), where N is the total number of entries
//param key
//return All the fields values contained into a hash.
func (r *Redis) Hvals(key string) ([]string, error) {
	err := r.client.hvals(key)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Return all the fields and associated values in a hash.
//
//<b>Time complexity:</b> O(N), where N is the total number of entries
//param key
//return All the fields and values contained into a hash.
func (r *Redis) HgetAll(key string) (map[string]string, error) {
	err := r.client.hgetAll(key)
	if err != nil {
		return nil, err
	}
	return StringArrayToMapReply(r.client.getMultiBulkReply())
}

//Add the string value to the head (LPUSH) or tail (RPUSH) of the list stored at key. If the key
//does not exist an empty list is created just before the append operation. If the key exists but
//is not a List an error is returned.
//
//Time complexity: O(1)
//param key
//param strings
//return Integer reply, specifically, the number of elements inside the list after the push
//        operation.
func (r *Redis) Rpush(key string, strings ...string) (int64, error) {
	err := r.client.rpush(key, strings...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Add the string value to the head (LPUSH) or tail (RPUSH) of the list stored at key. If the key
//does not exist an empty list is created just before the append operation. If the key exists but
//is not a List an error is returned.
//
//Time complexity: O(1)
//param key
//param strings
//return Integer reply, specifically, the number of elements inside the list after the push
//        operation.
func (r *Redis) Lpush(key string, strings ...string) (int64, error) {
	err := r.client.lpush(key, strings...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return the length of the list stored at the specified key. If the key does not exist zero is
//returned (the same behaviour as for empty lists). If the value stored at key is not a list an
//error is returned.
//
//Time complexity: O(1)
//param key
//return The length of the list.
func (r *Redis) Llen(key string) (int64, error) {
	err := r.client.llen(key)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return the specified elements of the list stored at the specified key. Start and end are
//zero-based indexes. 0 is the first element of the list (the list head), 1 the next element and
//so on.
//
//For example LRANGE foobar 0 2 will return the first three elements of the list.
//
//start and end can also be negative numbers indicating offsets from the end of the list. For
//example -1 is the last element of the list, -2 the penultimate element and so on.
//
//<b>Consistency with range functions in various programming languages</b>
//
//Note that if you have a list of numbers from 0 to 100, LRANGE 0 10 will return 11 elements,
//that is, rightmost item is included. This may or may not be consistent with behavior of
//range-related functions in your programming language of choice (think Ruby's Range.new,
//Array#slice or Python's range() function).
//
//LRANGE behavior is consistent with one of Tcl.
//
//<b>Out-of-range indexes</b>
//
//Indexes out of range will not produce an error: if start is over the end of the list, or start
//&gt; end, an empty list is returned. If end is over the end of the list Redis will threat it
//just like the last element of the list.
//
//Time complexity: O(start+n) (with n being the length of the range and start being the start
//offset)
//param key
//param start
//param end
//return Multi bulk reply, specifically a list of elements in the specified range.
func (r *Redis) Lrange(key string, start, stop int64) ([]string, error) {
	err := r.client.lrange(key, start, stop)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Trim an existing list so that it will contain only the specified range of elements specified.
//Start and end are zero-based indexes. 0 is the first element of the list (the list head), 1 the
//next element and so on.
//
//For example LTRIM foobar 0 2 will modify the list stored at foobar key so that only the first
//three elements of the list will remain.
//
//start and end can also be negative numbers indicating offsets from the end of the list. For
//example -1 is the last element of the list, -2 the penultimate element and so on.
//
//Indexes out of range will not produce an error: if start is over the end of the list, or start
//&gt; end, an empty list is left as value. If end over the end of the list Redis will threat it
//just like the last element of the list.
//
//Hint: the obvious use of LTRIM is together with LPUSH/RPUSH. For example:
//
//{@code lpush("mylist", "someelement"); ltrim("mylist", 0, 99); //}
//
//The above two commands will push elements in the list taking care that the list will not grow
//without limits. This is very useful when using Redis to store logs for example. It is important
//to note that when used in this way LTRIM is an O(1) operation because in the average case just
//one element is removed from the tail of the list.
//
//Time complexity: O(n) (with n being len of list - len of range)
//param key
//param start
//param end
//return Status code reply
func (r *Redis) Ltrim(key string, start, stop int64) (string, error) {
	err := r.client.ltrim(key, start, stop)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Return the specified element of the list stored at the specified key. 0 is the first element, 1
//the second and so on. Negative indexes are supported, for example -1 is the last element, -2
//the penultimate and so on.
//
//If the value stored at key is not of list type an error is returned. If the index is out of
//range a 'nil' reply is returned.
//
//Note that even if the average time complexity is O(n) asking for the first or the last element
//of the list is O(1).
//
//Time complexity: O(n) (with n being the length of the list)
//param key
//param index
//return Bulk reply, specifically the requested element
func (r *Redis) Lindex(key string, index int64) (string, error) {
	err := r.client.lindex(key, index)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Set a new value as the element at index position of the List at key.
//
//Out of range indexes will generate an error.
//
//Similarly to other list commands accepting indexes, the index can be negative to access
//elements starting from the end of the list. So -1 is the last element, -2 is the penultimate,
//and so forth.
//
//<b>Time complexity:</b>
//
//O(N) (with N being the length of the list), setting the first or last elements of the list is
//O(1).
//@see #lindex(String, long)
//param key
//param index
//param value
//return Status code reply
func (r *Redis) Lset(key string, index int64, value string) (string, error) {
	err := r.client.lset(key, index, value)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Remove the first count occurrences of the value element from the list. If count is zero all the
//elements are removed. If count is negative elements are removed from tail to head, instead to
//go from head to tail that is the normal behaviour. So for example LREM with count -2 and hello
//as value to remove against the list (a,b,c,hello,x,hello,hello) will lave the list
//(a,b,c,hello,x). The number of removed elements is returned as an integer, see below for more
//information about the returned value. Note that non existing keys are considered like empty
//lists by LREM, so LREM against non existing keys will always return 0.
//
//Time complexity: O(N) (with N being the length of the list)
//param key
//param count
//param value
//return Integer Reply, specifically: The number of removed elements if the operation succeeded
func (r *Redis) Lrem(key string, count int64, value string) (int64, error) {
	err := r.client.lrem(key, count, value)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Atomically return and remove the first (LPOP) or last (RPOP) element of the list. For example
//if the list contains the elements "a","b","c" LPOP will return "a" and the list will become
//"b","c".
//
//If the key does not exist or the list is already empty the special value 'nil' is returned.
//@see #rpop(String)
//param key
//return Bulk reply
func (r *Redis) Lpop(key string) (string, error) {
	err := r.client.lpop(key)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Atomically return and remove the first (LPOP) or last (RPOP) element of the list. For example
//if the list contains the elements "a","b","c" RPOP will return "c" and the list will become
//"a","b".
//
//If the key does not exist or the list is already empty the special value 'nil' is returned.
//@see #lpop(String)
//param key
//return Bulk reply
func (r *Redis) Rpop(key string) (string, error) {
	err := r.client.rpop(key)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Add the specified member to the set value stored at key. If member is already a member of the
//set no operation is performed. If key does not exist a new set with the specified member as
//sole member is created. If the key exists but does not hold a set value an error is returned.
//
//Time complexity O(1)
//param key
//param members
//return Integer reply, specifically: 1 if the new element was added 0 if the element was
//        already a member of the set
func (r *Redis) Sadd(key string, members ...string) (int64, error) {
	err := r.client.sadd(key, members...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return all the members (elements) of the set value stored at key. This is just syntax glue for
//{@link #sinter(String...) SINTER}.
//
//Time complexity O(N)
//param key
//return Multi bulk reply
func (r *Redis) Smembers(key string) ([]string, error) {
	err := r.client.smembers(key)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Remove the specified member from the set value stored at key. If member was not a member of the
//set no operation is performed. If key does not hold a set value an error is returned.
//
//Time complexity O(1)
//param key
//param members
//return Integer reply, specifically: 1 if the new element was removed 0 if the new element was
//        not a member of the set
func (r *Redis) Srem(key string, members ...string) (int64, error) {
	err := r.client.srem(key, members...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Remove a random element from a Set returning it as return value. If the Set is empty or the key
//does not exist, a nil object is returned.
//
//The {@link #srandmember(String)} command does a similar work but the returned element is not
//removed from the Set.
//
//Time complexity O(1)
//param key
//return Bulk reply
func (r *Redis) Spop(key string) (string, error) {
	err := r.client.spop(key)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

// remove multi random element
// see Spop(key string)
func (r *Redis) SpopBatch(key string, count int64) ([]string, error) {
	err := r.client.spopBatch(key, count)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Return the set cardinality (number of elements). If the key does not exist 0 is returned, like
//for empty sets.
//param key
//return Integer reply, specifically: the cardinality (number of elements) of the set as an
//        integer.
func (r *Redis) Scard(key string) (int64, error) {
	err := r.client.scard(key)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return 1 if member is a member of the set stored at key, otherwise 0 is returned.
//
//Time complexity O(1)
//param key
//param member
//return Integer reply, specifically: 1 if the element is a member of the set 0 if the element
//        is not a member of the set OR if the key does not exist
func (r *Redis) Sismember(key, member string) (bool, error) {
	err := r.client.sismember(key, member)
	if err != nil {
		return false, err
	}
	return Int64ToBoolReply(r.client.getIntegerReply())
}

//Return the members of a set resulting from the intersection of all the sets hold at the
//specified keys. Like in {@link #lrange(String, long, long) LRANGE} the result is sent to the
//client as a multi-bulk reply (see the protocol specification for more information). If just a
//single key is specified, then this command produces the same result as
//{@link #smembers(String) SMEMBERS}. Actually SMEMBERS is just syntax sugar for SINTER.
//
//Non existing keys are considered like empty sets, so if one of the keys is missing an empty set
//is returned (since the intersection with an empty set always is an empty set).
//
//Time complexity O(N*M) worst case where N is the cardinality of the smallest set and M the
//number of sets
//param keys
//return Multi bulk reply, specifically the list of common elements.
func (r *Redis) Sinter(keys ...string) ([]string, error) {
	err := r.client.sinter(keys...)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//This commnad works exactly like {@link #sinter(String...) SINTER} but instead of being returned
//the resulting set is sotred as dstkey.
//
//Time complexity O(N*M) worst case where N is the cardinality of the smallest set and M the
//number of sets
//param dstkey
//param keys
//return Status code reply
func (r *Redis) Sinterstore(dstkey string, keys ...string) (int64, error) {
	err := r.client.sinterstore(dstkey, keys...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return the members of a set resulting from the union of all the sets hold at the specified
//keys. Like in {@link #lrange(String, long, long) LRANGE} the result is sent to the client as a
//multi-bulk reply (see the protocol specification for more information). If just a single key is
//specified, then this command produces the same result as {@link #smembers(String) SMEMBERS}.
//
//Non existing keys are considered like empty sets.
//
//Time complexity O(N) where N is the total number of elements in all the provided sets
//param keys
//return Multi bulk reply, specifically the list of common elements.
func (r *Redis) Sunion(keys ...string) ([]string, error) {
	err := r.client.sunion(keys...)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//This command works exactly like {@link #sunion(String...) SUNION} but instead of being returned
//the resulting set is stored as dstkey. Any existing value in dstkey will be over-written.
//
//Time complexity O(N) where N is the total number of elements in all the provided sets
//param dstkey
//param keys
//return Status code reply
func (r *Redis) Sunionstore(dstkey string, keys ...string) (int64, error) {
	err := r.client.sunionstore(dstkey, keys...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return the difference between the Set stored at key1 and all the Sets key2, ..., keyN
//
//<b>Example:</b>
//<pre>
//key1 = [x, a, b, c]
//key2 = [c]
//key3 = [a, d]
//SDIFF key1,key2,key3 =&gt; [x, b]
//</pre>
//Non existing keys are considered like empty sets.
//
//<b>Time complexity:</b>
//
//O(N) with N being the total number of elements of all the sets
//param keys
//return Return the members of a set resulting from the difference between the first set
//        provided and all the successive sets.
func (r *Redis) Sdiff(keys ...string) ([]string, error) {
	err := r.client.sdiff(keys...)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//This command works exactly like {@link #sdiff(String...) SDIFF} but instead of being returned
//the resulting set is stored in dstkey.
//param dstkey
//param keys
//return Status code reply
func (r *Redis) Sdiffstore(dstkey string, keys ...string) (int64, error) {
	err := r.client.sdiffstore(dstkey, keys...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return a random element from a Set, without removing the element. If the Set is empty or the
//key does not exist, a nil object is returned.
//
//The SPOP command does a similar work but the returned element is popped (removed) from the Set.
//
//Time complexity O(1)
//param key
//return Bulk reply
func (r *Redis) Srandmember(key string) (string, error) {
	err := r.client.srandmember(key)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Add the specified member having the specifeid score to the sorted set stored at key. If member
//is already a member of the sorted set the score is updated, and the element reinserted in the
//right position to ensure sorting. If key does not exist a new sorted set with the specified
//member as sole member is crated. If the key exists but does not hold a sorted set value an
//error is returned.
//
//The score value can be the string representation of a double precision floating point number.
//
//Time complexity O(log(N)) with N being the number of elements in the sorted set
//param key
//param score
//param member
//return Integer reply, specifically: 1 if the new element was added 0 if the element was
//        already a member of the sorted set and the score was updated
func (r *Redis) Zadd(key string, score float64, member string, mparams ...ZAddParams) (int64, error) {
	err := r.client.zadd(key, score, member)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Returns the specified range of elements in the sorted set stored at key.
// The elements are considered to be ordered from the lowest to the highest score.
// Lexicographical order is used for elements with equal score.
//See ZREVRANGE when you need the elements ordered from highest to lowest score
// (and descending lexicographical order for elements with equal score).
//Both start and stop are zero-based indexes, where 0 is the first element,
// 1 is the next element and so on. They can also be negative numbers indicating offsets from the end of the sorted set,
// with -1 being the last element of the sorted set, -2 the penultimate element and so on.
//start and stop are inclusive ranges,
// so for example ZRANGE myzset 0 1 will return both the first and the second element of the sorted set.
//Out of range indexes will not produce an error. If start is larger than the largest index in the sorted set,
// or start > stop, an empty list is returned. If stop is larger than the end of the sorted set Redis will treat it
// like it is the last element of the sorted set.
//It is possible to pass the WITHSCORES option in order to return the scores of the elements together with the elements.
// The returned list will contain value1,score1,...,valueN,scoreN instead of value1,...,valueN.
// Client libraries are free to return a more appropriate data type (suggestion: an array with (value, score) arrays/tuples).
//Return value
//Array reply: list of elements in the specified range (optionally with their scores, in case the WITHSCORES option is given).
func (r *Redis) Zrange(key string, start, stop int64) ([]string, error) {
	err := r.client.zrange(key, start, stop)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Remove the specified member from the sorted set value stored at key. If member was not a member
//of the set no operation is performed. If key does not not hold a set value an error is
//returned.
//
//Time complexity O(log(N)) with N being the number of elements in the sorted set
//param key
//param members
//return Integer reply, specifically: 1 if the new element was removed 0 if the new element was
//        not a member of the set
func (r *Redis) Zrem(key string, members ...string) (int64, error) {
	err := r.client.zrem(key, members...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//If member already exists in the sorted set adds the increment to its score and updates the
//position of the element in the sorted set accordingly. If member does not already exist in the
//sorted set it is added with increment as score (that is, like if the previous score was
//virtually zero). If key does not exist a new sorted set with the specified member as sole
//member is crated. If the key exists but does not hold a sorted set value an error is returned.
//
//The score value can be the string representation of a double precision floating point number.
//It's possible to provide a negative value to perform a decrement.
//
//For an introduction to sorted sets check the Introduction to Redis data types page.
//
//Time complexity O(log(N)) with N being the number of elements in the sorted set
//param key
//param score
//param member
//return The new score
func (r *Redis) Zincrby(key string, increment float64, member string, params ...ZAddParams) (float64, error) {
	err := r.client.zincrby(key, increment, member)
	if err != nil {
		return 0, err
	}
	return StringToFloat64Reply(r.client.getBulkReply())
}

//Return the rank (or index) or member in the sorted set at key, with scores being ordered from
//low to high.
//
//When the given member does not exist in the sorted set, the special value 'nil' is returned.
//The returned rank (or index) of the member is 0-based for both commands.
//
//<b>Time complexity:</b>
//
//O(log(N))
//@see #zrevrank(String, String)
//param key
//param member
//return Integer reply or a nil bulk reply, specifically: the rank of the element as an integer
//        reply if the element exists. A nil bulk reply if there is no such element.
func (r *Redis) Zrank(key, member string) (int64, error) {
	err := r.client.zrank(key, member)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return the rank (or index) or member in the sorted set at key, with scores being ordered from
//high to low.
//
//When the given member does not exist in the sorted set, the special value 'nil' is returned.
//The returned rank (or index) of the member is 0-based for both commands.
//
//<b>Time complexity:</b>
//
//O(log(N))
//@see #zrank(String, String)
//param key
//param member
//return Integer reply or a nil bulk reply, specifically: the rank of the element as an integer
//        reply if the element exists. A nil bulk reply if there is no such element.
func (r *Redis) Zrevrank(key, member string) (int64, error) {
	err := r.client.zrevrank(key, member)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Returns the specified range of elements in the sorted set stored at key.
// The elements are considered to be ordered from the highest to the lowest score.
// Descending lexicographical order is used for elements with equal score.
//Apart from the reversed ordering, ZREVRANGE is similar to ZRANGE.
//Return value
//Array reply: list of elements in the specified range (optionally with their scores).
func (r *Redis) Zrevrange(key string, start, stop int64) ([]string, error) {
	err := r.client.zrevrange(key, start, stop)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Return the sorted set cardinality (number of elements). If the key does not exist 0 is
//returned, like for empty sorted sets.
//
//Time complexity O(1)
//param key
//return the cardinality (number of elements) of the set as an integer.
func (r *Redis) Zcard(key string) (int64, error) {
	err := r.client.zcard(key)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return the score of the specified element of the sorted set at key. If the specified element
//does not exist in the sorted set, or the key does not exist at all, a special 'nil' value is
//returned.
//
//<b>Time complexity:</b> O(1)
//param key
//param member
//return the score
func (r *Redis) Zscore(key, member string) (float64, error) {
	err := r.client.zscore(key, member)
	if err != nil {
		return 0, err
	}
	return StringToFloat64Reply(r.client.getBulkReply())
}

//Marks the given keys to be watched for conditional execution of a transaction.
//
//Return value
//Simple string reply: always OK.
func (r *Redis) Watch(keys ...string) (string, error) {
	err := r.client.watch(keys...)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//Sort a Set or a List.
//
//Sort the elements contained in the List, Set, or Sorted Set value at key. By default sorting is
//numeric with elements being compared as double precision floating point numbers. This is the
//simplest form of SORT.
//@see #sort(String, String)
//@see #sort(String, SortingParams)
//@see #sort(String, SortingParams, String)
//param key
//return Assuming the Set/List at key contains a list of numbers, the return value will be the
//        list of numbers ordered from the smallest to the biggest number.
func (r *Redis) Sort(key string, sortingParameters ...SortingParams) ([]string, error) {
	err := r.client.sort(key, sortingParameters...)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

// Zcount Returns the number of elements in the sorted set at key with a score between min and max.
// The min and max arguments have the same semantic as described for ZRANGEBYSCORE.
// Note: the command has a complexity of just O(log(N))
// because it uses elements ranks (see ZRANK) to get an idea of the range.
// Because of this there is no need to do a work proportional to the size of the range.
//
// Return value
// Integer reply: the number of elements in the specified score range.
func (r *Redis) Zcount(key, min, max string) (int64, error) {
	err := r.client.zcount(key, min, max)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Return the all the elements in the sorted set at key with a score between min and max
//(including elements with score equal to min or max).
//
//The elements having the same score are returned sorted lexicographically as ASCII strings (this
//follows from a property of Redis sorted sets and does not involve further computation).
//
//Using the optional {@link #zrangeByScore(String, double, double, int, int) LIMIT} it's possible
//to get only a range of the matching elements in an SQL-alike way. Note that if offset is large
//the commands needs to traverse the list for offset elements and this adds up to the O(M)
//figure.
//
//The {@link #zcount(String, double, double) ZCOUNT} command is similar to
//{@link #zrangeByScore(String, double, double) ZRANGEBYSCORE} but instead of returning the
//actual elements in the specified interval, it just returns the number of matching elements.
//
//<b>Exclusive intervals and infinity</b>
//
//min and max can be -inf and +inf, so that you are not required to know what's the greatest or
//smallest element in order to take, for instance, elements "up to a given value".
//
//Also while the interval is for default closed (inclusive) it's possible to specify open
//intervals prefixing the score with a "(" character, so for instance:
//
//{@code ZRANGEBYSCORE zset (1.3 5}
//
//Will return all the values with score &gt; 1.3 and &lt;= 5, while for instance:
//
//{@code ZRANGEBYSCORE zset (5 (10}
//
//Will return all the values with score &gt; 5 and &lt; 10 (5 and 10 excluded).
//
//<b>Time complexity:</b>
//
//O(log(N))+O(M) with N being the number of elements in the sorted set and M the number of
//elements returned by the command, so if M is constant (for instance you always ask for the
//first ten elements with LIMIT) you can consider it O(log(N))
//@see #zrangeByScore(String, double, double)
//@see #zrangeByScore(String, double, double, int, int)
//@see #zrangeByScoreWithScores(String, double, double)
//@see #zrangeByScoreWithScores(String, String, String)
//@see #zrangeByScoreWithScores(String, double, double, int, int)
//@see #zcount(String, double, double)
//param key
//param min a double or Double.MIN_VALUE for "-inf"
//param max a double or Double.MAX_VALUE for "+inf"
//return Multi bulk reply specifically a list of elements in the specified score range.
func (r *Redis) ZrangeByScore(key, min, max string) ([]string, error) {
	err := r.client.zrangeByScore(key, min, max)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Return the all the elements in the sorted set at key with a score between min and max
//(including elements with score equal to min or max).
//
//The elements having the same score are returned sorted lexicographically as ASCII strings (this
//follows from a property of Redis sorted sets and does not involve further computation).
//
//Using the optional {@link #zrangeByScore(String, double, double, int, int) LIMIT} it's possible
//to get only a range of the matching elements in an SQL-alike way. Note that if offset is large
//the commands needs to traverse the list for offset elements and this adds up to the O(M)
//figure.
//
//The {@link #zcount(String, double, double) ZCOUNT} command is similar to
//{@link #zrangeByScore(String, double, double) ZRANGEBYSCORE} but instead of returning the
//actual elements in the specified interval, it just returns the number of matching elements.
//
//<b>Exclusive intervals and infinity</b>
//
//min and max can be -inf and +inf, so that you are not required to know what's the greatest or
//smallest element in order to take, for instance, elements "up to a given value".
//
//Also while the interval is for default closed (inclusive) it's possible to specify open
//intervals prefixing the score with a "(" character, so for instance:
//
//{@code ZRANGEBYSCORE zset (1.3 5}
//
//Will return all the values with score &gt; 1.3 and &lt;= 5, while for instance:
//
//{@code ZRANGEBYSCORE zset (5 (10}
//
//Will return all the values with score &gt; 5 and &lt; 10 (5 and 10 excluded).
//
//<b>Time complexity:</b>
//
//O(log(N))+O(M) with N being the number of elements in the sorted set and M the number of
//elements returned by the command, so if M is constant (for instance you always ask for the
//first ten elements with LIMIT) you can consider it O(log(N))
//@see #zrangeByScore(String, double, double)
//@see #zrangeByScore(String, double, double, int, int)
//@see #zrangeByScoreWithScores(String, double, double)
//@see #zrangeByScoreWithScores(String, double, double, int, int)
//@see #zcount(String, double, double)
//param key
//param min
//param max
//return Multi bulk reply specifically a list of elements in the specified score range.
func (r *Redis) ZrangeByScoreWithScores(key, min, max string) ([]Tuple, error) {
	panic("not implement!")
}

// ZrevrangeByScore Returns all the elements in the sorted set at key with a score between max and min
// (including elements with score equal to max or min). In contrary to the default ordering of sorted sets,
// for this command the elements are considered to be ordered from high to low scores.
// The elements having the same score are returned in reverse lexicographical order.
// Apart from the reversed ordering, ZREVRANGEBYSCORE is similar to ZRANGEBYSCORE.
//
// Return value
// Array reply: list of elements in the specified score range (optionally with their scores).
func (r *Redis) ZrevrangeByScore(key, max, min string) ([]string, error) {
	err := r.client.zrevrangeByScore(key, max, min)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

// see ZrevrangeByScore(key, max, min string)
func (r *Redis) ZrevrangeByScoreWithScores(key, max, min string) ([]Tuple, error) {
	err := r.client.zrevrangeByScore(key, max, min)
	if err != nil {
		return nil, err
	}
	return StringArrToTupleReply(r.client.getMultiBulkReply())
}

//Remove all elements in the sorted set at key with rank between start and end. Start and end are
//0-based with rank 0 being the element with the lowest score. Both start and end can be negative
//numbers, where they indicate offsets starting at the element with the highest rank. For
//example: -1 is the element with the highest score, -2 the element with the second highest score
//and so forth.
//
//<b>Time complexity:</b> O(log(N))+O(M) with N being the number of elements in the sorted set
//and M the number of elements removed by the operation
func (r *Redis) ZremrangeByRank(key string, start, stop int64) (int64, error) {
	err := r.client.zremrangeByRank(key, start, stop)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

// Strlen Returns the length of the string value stored at key.
// An error is returned when key holds a non-string value.
// Return value
// Integer reply: the length of the string at key, or 0 when key does not exist.
func (r *Redis) Strlen(key string) (int64, error) {
	err := r.client.strlen(key)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

// Lpushx Inserts value at the head of the list stored at key,
// only if key already exists and holds a list. In contrary to LPUSH,
// no operation will be performed when key does not yet exist.
// Return value
// Integer reply: the length of the list after the push operation.
func (r *Redis) Lpushx(key string, string ...string) (int64, error) {
	err := r.client.lpushx(key, string...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Undo a {@link #expire(String, int) expire} at turning the expire key into a normal key.
//
//Time complexity: O(1)
//param key
//return Integer reply, specifically: 1: the key is now persist. 0: the key is not persist (only
//        happens when key not set).
func (r *Redis) Persist(key string) (int64, error) {
	err := r.client.persist(key)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

// Rpushx Inserts value at the tail of the list stored at key,
// only if key already exists and holds a list. In contrary to RPUSH,
// no operation will be performed when key does not yet exist.
//
// Return value
// Integer reply: the length of the list after the push operation.
func (r *Redis) Rpushx(key string, string ...string) (int64, error) {
	err := r.client.rpushx(key, string...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

// Echo Returns message.
//
// Return value
// Bulk string reply
func (r *Redis) Echo(string string) (string, error) {
	err := r.client.echo(string)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

// see SetWithParamsAndTime(key, value, nxxx, expx string, time int64)
func (r *Redis) SetWithParams(key, value, nxxx string) (string, error) {
	err := r.client.setWithParams(key, value, nxxx)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//This command works exactly like EXPIRE but the time to live of the key is specified in milliseconds instead of seconds.
//
//Return value
//Integer reply, specifically:
//
//1 if the timeout was set.
//0 if key does not exist.
func (r *Redis) Pexpire(key string, milliseconds int64) (int64, error) {
	err := r.client.pexpire(key, milliseconds)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//PEXPIREAT has the same effect and semantic as EXPIREAT,
// but the Unix time at which the key will expire is specified in milliseconds instead of seconds.
//
//Return value
//Integer reply, specifically:
//
//1 if the timeout was set.
//0 if key does not exist.
func (r *Redis) PexpireAt(key string, millisecondsTimestamp int64) (int64, error) {
	err := r.client.pexpireAt(key, millisecondsTimestamp)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

// see Setbit(key string, offset int64, value string)
func (r *Redis) SetbitWithBool(key string, offset int64, value bool) (bool, error) {
	var valueByte []byte
	if value {
		valueByte = BytesTrue
	} else {
		valueByte = BytesFalse
	}
	return r.Setbit(key, offset, string(valueByte))
}

//Sets or clears the bit at offset in the string value stored at key.
//
//The bit is either set or cleared depending on value, which can be either 0 or 1.
// When key does not exist, a new string value is created.
// The string is grown to make sure it can hold a bit at offset.
// The offset argument is required to be greater than or equal to 0,
// and smaller than 232 (this limits bitmaps to 512MB). When the string at key is grown, added bits are set to 0.
//
//Warning: When setting the last possible bit (offset equal to 232 -1) and
// the string value stored at key does not yet hold a string value,
// or holds a small string value,
// Redis needs to allocate all intermediate memory which can block the server for some time.
// On a 2010 MacBook Pro, setting bit number 232 -1 (512MB allocation) takes ~300ms,
// setting bit number 230 -1 (128MB allocation) takes ~80ms,
// setting bit number 228 -1 (32MB allocation) takes ~30ms and setting bit number 226 -1 (8MB allocation) takes ~8ms.
// Note that once this first allocation is done,
// subsequent calls to SETBIT for the same key will not have the allocation overhead.
//
//Return value
//Integer reply: the original bit value stored at offset.
func (r *Redis) Setbit(key string, offset int64, value string) (bool, error) {
	err := r.client.setbit(key, offset, value)
	if err != nil {
		return false, err
	}
	return Int64ToBoolReply(r.client.getIntegerReply())
}

//Returns the bit value at offset in the string value stored at key.
//
//When offset is beyond the string length,
// the string is assumed to be a contiguous space with 0 bits.
// When key does not exist it is assumed to be an empty string,
// so offset is always out of range and the value is also assumed to be a contiguous space with 0 bits.
//
//Return value
//Integer reply: the bit value stored at offset.
func (r *Redis) Getbit(key string, offset int64) (bool, error) {
	err := r.client.getbit(key, offset)
	if err != nil {
		return false, err
	}
	return Int64ToBoolReply(r.client.getIntegerReply())
}

// PSETEX works exactly like SETEX with the sole difference that the expire time is specified in milliseconds instead of seconds.
func (r *Redis) Psetex(key string, milliseconds int64, value string) (string, error) {
	err := r.client.psetex(key, milliseconds, value)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

// see Srandmember(key string)
func (r *Redis) SrandmemberBatch(key string, count int) ([]string, error) {
	err := r.client.srandmemberBatch(key, count)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//see Zadd(key string, score float64, member string, mparams ...ZAddParams)
func (r *Redis) ZaddByMap(key string, scoreMembers map[string]float64, params ...ZAddParams) (int64, error) {
	err := r.client.zaddByMap(key, scoreMembers, params...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//see Zrange()
func (r *Redis) ZrangeWithScores(key string, start, end int64) ([]Tuple, error) {
	err := r.client.zrangeWithScores(key, start, end)
	if err != nil {
		return nil, err
	}
	return StringArrToTupleReply(r.client.getMultiBulkReply())
}

//see Zrevrange()
func (r *Redis) ZrevrangeWithScores(key string, start, end int64) ([]Tuple, error) {
	err := r.client.zrevrangeWithScores(key, start, end)
	if err != nil {
		return nil, err
	}
	return StringArrToTupleReply(r.client.getMultiBulkReply())
}

//see Zrange()
func (r *Redis) ZrangeByScoreBatch(key, min, max string, offset, count int) ([]string, error) {
	err := r.client.zrangeByScoreBatch(key, min, max, offset, count)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//see Zrange()
func (r *Redis) ZrangeByScoreWithScoresBatch(key, min, max string, offset, count int) ([]Tuple, error) {
	err := r.client.zrangeByScoreBatch(key, min, max, offset, count)
	if err != nil {
		return nil, err
	}
	return StringArrToTupleReply(r.client.getMultiBulkReply())
}

//see Zrevrange()
func (r *Redis) ZrevrangeByScoreWithScoresBatch(key, max, min string, offset, count int) ([]Tuple, error) {
	err := r.client.zrevrangeByScoreBatch(key, max, min, offset, count)
	if err != nil {
		return nil, err
	}
	return StringArrToTupleReply(r.client.getMultiBulkReply())
}

//Removes all elements in the sorted set stored at key with a score between min and max (inclusive).
//
//Since version 2.1.6, min and max can be exclusive, following the syntax of ZRANGEBYSCORE.
//
//Return value
//Integer reply: the number of elements removed.
func (r *Redis) ZremrangeByScore(key, start, end string) (int64, error) {
	err := r.client.zremrangeByScore(key, start, end)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//When all the elements in a sorted set are inserted with the same score,
// in order to force lexicographical ordering,
// this command returns the number of elements in the sorted set at key with a value between min and max.
//
//The min and max arguments have the same meaning as described for ZRANGEBYLEX.
//
//Note: the command has a complexity of just O(log(N))
// because it uses elements ranks (see ZRANK) to get an idea of the range.
// Because of this there is no need to do a work proportional to the size of the range.
//
//Return value
//Integer reply: the number of elements in the specified score range.
func (r *Redis) Zlexcount(key, min, max string) (int64, error) {
	err := r.client.zlexcount(key, min, max)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//When all the elements in a sorted set are inserted with the same score,
// in order to force lexicographical ordering,
// this command returns all the elements in the sorted set at key with a value between min and max.
//
//If the elements in the sorted set have different scores, the returned elements are unspecified.
//
//The elements are considered to be ordered from lower to higher strings as compared byte-by-byte
// using the memcmp() C function. Longer strings are considered greater than shorter strings if the common part is identical.
//
//The optional LIMIT argument can be used to only get a range of the matching elements
// (similar to SELECT LIMIT offset, count in SQL).
// A negative count returns all elements from the offset.
// Keep in mind that if offset is large, the sorted set needs to be traversed
// for offset elements before getting to the elements to return, which can add up to O(N) time complexity.
//Return value
//Array reply: list of elements in the specified score range.
func (r *Redis) ZrangeByLex(key, min, max string) ([]string, error) {
	err := r.client.zrangeByLex(key, min, max)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//see ZrangeByLex()
func (r *Redis) ZrangeByLexBatch(key, min, max string, offset, count int) ([]string, error) {
	err := r.client.zrangeByLexBatch(key, min, max, offset, count)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//When all the elements in a sorted set are inserted with the same score,
// in order to force lexicographical ordering,
// this command returns all the elements in the sorted set at key with a value between max and min.
//
//Apart from the reversed ordering, ZREVRANGEBYLEX is similar to ZRANGEBYLEX.
//
//Return value
//Array reply: list of elements in the specified score range.
func (r *Redis) ZrevrangeByLex(key, max, min string) ([]string, error) {
	err := r.client.zrevrangeByLex(key, max, min)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

// see ZrevrangeByLex()
func (r *Redis) ZrevrangeByLexBatch(key, max, min string, offset, count int) ([]string, error) {
	err := r.client.zrevrangeByLexBatch(key, max, min, offset, count)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//When all the elements in a sorted set are inserted with the same score,
// in order to force lexicographical ordering,
// this command removes all elements in the sorted set stored at key
// between the lexicographical range specified by min and max.
//
//The meaning of min and max are the same of the ZRANGEBYLEX command.
// Similarly, this command actually returns the same elements that ZRANGEBYLEX would return
// if called with the same min and max arguments.
//
//Return value
//Integer reply: the number of elements removed.
func (r *Redis) ZremrangeByLex(key, min, max string) (int64, error) {
	err := r.client.zremrangeByLex(key, min, max)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Inserts value in the list stored at key either before or after the reference value pivot.
//
//When key does not exist, it is considered an empty list and no operation is performed.
//
//An error is returned when key exists but does not hold a list value.
//
//Return value
//Integer reply: the length of the list after the insert operation, or -1 when the value pivot was not found.
func (r *Redis) Linsert(key string, where ListOption, pivot, value string) (int64, error) {
	err := r.client.linsert(key, where, pivot, value)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Move key from the currently selected database (see SELECT) to the specified destination database.
// When key already exists in the destination database,
// or it does not exist in the source database, it does nothing.
// It is possible to use MOVE as a locking primitive because of this.
//
//Return value
//Integer reply, specifically:
//
//1 if key was moved.
//0 if key was not moved.
func (r *Redis) Move(key string, dbIndex int) (int64, error) {
	err := r.client.move(key, dbIndex)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Count the number of set bits (population counting) in a string.
//
//By default all the bytes contained in the string are examined.
// It is possible to specify the counting operation only in an interval passing the additional arguments start and end.
//
//Like for the GETRANGE command start and end can contain negative values
// in order to index bytes starting from the end of the string, where -1 is the last byte, -2 is the penultimate, and so forth.
//
//Non-existent keys are treated as empty strings, so the command will return zero.
//
//Return value
//Integer reply
//
//The number of bits set to 1.
func (r *Redis) Bitcount(key string) (int64, error) {
	err := r.client.bitcount(key)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

// see Bitcount()
func (r *Redis) BitcountRange(key string, start, end int64) (int64, error) {
	err := r.client.bitcountRange(key, start, end)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Bitpos Return the position of the first bit set to 1 or 0 in a string.
func (r *Redis) Bitpos(key string, value bool, params ...BitPosParams) (int64, error) {
	err := r.client.bitpos(key, value, params...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Hscan ...
func (r *Redis) Hscan(key, cursor string, params ...ScanParams) (*ScanResult, error) {
	err := r.client.hscan(key, cursor, params...)
	if err != nil {
		return nil, err
	}
	return ObjectArrToScanResultReply(r.client.getObjectMultiBulkReply())
}

//Sscan ...
func (r *Redis) Sscan(key, cursor string, params ...ScanParams) (*ScanResult, error) {
	err := r.client.sscan(key, cursor, params...)
	if err != nil {
		return nil, err
	}
	return ObjectArrToScanResultReply(r.client.getObjectMultiBulkReply())
}

//Zscan ...
func (r *Redis) Zscan(key, cursor string, params ...ScanParams) (*ScanResult, error) {
	err := r.client.zscan(key, cursor, params...)
	if err != nil {
		return nil, err
	}
	return ObjectArrToScanResultReply(r.client.getObjectMultiBulkReply())
}

//Pfadd ...
func (r *Redis) Pfadd(key string, elements ...string) (int64, error) {
	err := r.client.pfadd(key, elements...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Geoadd ...
func (r *Redis) Geoadd(key string, longitude, latitude float64, member string) (int64, error) {
	err := r.client.geoadd(key, longitude, latitude, member)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//GeoaddByMap ...
func (r *Redis) GeoaddByMap(key string, memberCoordinateMap map[string]GeoCoordinate) (int64, error) {
	err := r.client.geoaddByMap(key, memberCoordinateMap)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Geodist ...
func (r *Redis) Geodist(key, member1, member2 string, unit ...GeoUnit) (float64, error) {
	err := r.client.geodist(key, member1, member2, unit...)
	if err != nil {
		return 0, err
	}
	return StringToFloat64Reply(r.client.getBulkReply())
}

//Geohash ...
func (r *Redis) Geohash(key string, members ...string) ([]string, error) {
	err := r.client.geohash(key, members...)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Geopos ...
func (r *Redis) Geopos(key string, members ...string) ([]*GeoCoordinate, error) {
	err := r.client.geopos(key, members...)
	if err != nil {
		return nil, err
	}
	return ObjectArrToGeoCoordinateReply(r.client.getObjectMultiBulkReply())
}

//Georadius ...
func (r *Redis) Georadius(key string, longitude, latitude, radius float64, unit GeoUnit, param ...GeoRadiusParam) ([]*GeoCoordinate, error) {
	err := r.client.georadius(key, longitude, latitude, radius, unit, param...)
	if err != nil {
		return nil, err
	}
	return ObjectArrToGeoCoordinateReply(r.client.getObjectMultiBulkReply())
}

//GeoradiusByMember ...
func (r *Redis) GeoradiusByMember(key, member string, radius float64, unit GeoUnit, param ...GeoRadiusParam) ([]*GeoCoordinate, error) {
	err := r.client.georadiusByMember(key, member, radius, unit, param...)
	if err != nil {
		return nil, err
	}
	return ObjectArrToGeoCoordinateReply(r.client.getObjectMultiBulkReply())
}

//Bitfield The command treats a Redis string as a array of bits,
// and is capable of addressing specific integer fields of varying bit widths and arbitrary non (necessary) aligned offset.
func (r *Redis) Bitfield(key string, arguments ...string) ([]int64, error) {
	err := r.client.bitfield(key, arguments...)
	if err != nil {
		return nil, err
	}
	return r.client.getIntegerMultiBulkReply()
}

//</editor-fold>

//<editor-fold desc="multikeycommands">

//Returns all the keys matching the glob-style pattern as space separated strings. For example if
//you have in the database the keys "foo" and "foobar" the command "KEYS foo*" will return
//"foo foobar".
//
//Note that while the time complexity for this operation is O(n) the constant times are pretty
//low. For example Redis running on an entry level laptop can scan a 1 million keys database in
//40 milliseconds. <b>Still it's better to consider this one of the slow commands that may ruin
//the DB performance if not used with care.</b>
//
//In other words this command is intended only for debugging and special operations like creating
//a script to change the DB schema. Don't use it in your normal code. Use Redis Sets in order to
//group together a subset of objects.
//
//Glob style patterns examples:
//<ul>
//<li>h?llo will match hello hallo hhllo
//<li>h*llo will match hllo heeeello
//<li>h[ae]llo will match hello and hallo, but not hillo
//</ul>
//
//Use \ to escape special chars if you want to match them verbatim.
//
//Time complexity: O(n) (with n being the number of keys in the DB, and assuming keys and pattern
//of limited length)
//param pattern
//return Multi bulk reply
func (r *Redis) Keys(pattern string) ([]string, error) {
	err := r.client.keys(pattern)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Remove the specified keys. If a given key does not exist no operation is performed for this
//key. The command returns the number of keys removed. Time complexity: O(1)
//param keys
//return Integer reply, specifically: an integer greater than 0 if one or more keys were removed
//        0 if none of the specified key existed
func (r *Redis) Del(key ...string) (int64, error) {
	err := r.client.del(key...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Test if the specified key exists. The command returns the number of keys existed Time
//complexity: O(N)
//param keys
//return Integer reply, specifically: an integer greater than 0 if one or more keys were removed
//        0 if none of the specified key existed
func (r *Redis) Exists(keys ...string) (int64, error) {
	err := r.client.exists(keys...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Atomically renames the key oldkey to newkey. If the source and destination name are the same an
//error is returned. If newkey already exists it is overwritten.
//
//Time complexity: O(1)
//param oldkey
//param newkey
//return Status code repy
func (r *Redis) Rename(oldkey, newkey string) (string, error) {
	err := r.client.rename(oldkey, newkey)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//Rename oldkey into newkey but fails if the destination key newkey already exists.
//
//Time complexity: O(1)
//param oldkey
//param newkey
//return Integer reply, specifically: 1 if the key was renamed 0 if the target key already exist
func (r *Redis) Renamenx(oldkey, newkey string) (int64, error) {
	err := r.client.renamenx(oldkey, newkey)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Get the values of all the specified keys. If one or more keys dont exist or is not of type
//String, a 'nil' value is returned instead of the value of the specified key, but the operation
//never fails.
//
//Time complexity: O(1) for every key
//param keys
//return Multi bulk reply
func (r *Redis) Mget(keys ...string) ([]string, error) {
	err := r.client.mget(keys...)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Set the the respective keys to the respective values. MSET will replace old values with new
//values, while {@link #msetnx(String...) MSETNX} will not perform any operation at all even if
//just a single key already exists.
//
//Because of this semantic MSETNX can be used in order to set different keys representing
//different fields of an unique logic object in a way that ensures that either all the fields or
//none at all are set.
//
//Both MSET and MSETNX are atomic operations. This means that for instance if the keys A and B
//are modified, another client talking to Redis can either see the changes to both A and B at
//once, or no modification at all.
//@see #msetnx(String...)
//param keysvalues
//return Status code reply Basically +OK as MSET can't fail
func (r *Redis) Mset(keysvalues ...string) (string, error) {
	err := r.client.mset(keysvalues...)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Set the the respective keys to the respective values. {@link #mset(String...) MSET} will
//replace old values with new values, while MSETNX will not perform any operation at all even if
//just a single key already exists.
//
//Because of this semantic MSETNX can be used in order to set different keys representing
//different fields of an unique logic object in a way that ensures that either all the fields or
//none at all are set.
//
//Both MSET and MSETNX are atomic operations. This means that for instance if the keys A and B
//are modified, another client talking to Redis can either see the changes to both A and B at
//once, or no modification at all.
//@see #mset(String...)
//param keysvalues
//return Integer reply, specifically: 1 if the all the keys were set 0 if no key was set (at
//        least one key already existed)
func (r *Redis) Msetnx(keysvalues ...string) (int64, error) {
	err := r.client.msetnx(keysvalues...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Atomically return and remove the last (tail) element of the srckey list, and push the element
//as the first (head) element of the dstkey list. For example if the source list contains the
//elements "a","b","c" and the destination list contains the elements "foo","bar" after an
//RPOPLPUSH command the content of the two lists will be "a","b" and "c","foo","bar".
//
//If the key does not exist or the list is already empty the special value 'nil' is returned. If
//the srckey and dstkey are the same the operation is equivalent to removing the last element
//from the list and pusing it as first element of the list, so it's a "list rotation" command.
//
//Time complexity: O(1)
//param srckey
//param dstkey
//return Bulk reply
func (r *Redis) Rpoplpush(srckey, dstkey string) (string, error) {
	err := r.client.rpopLpush(srckey, dstkey)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Move the specifided member from the set at srckey to the set at dstkey. This operation is
//atomic, in every given moment the element will appear to be in the source or destination set
//for accessing clients.
//
//If the source set does not exist or does not contain the specified element no operation is
//performed and zero is returned, otherwise the element is removed from the source set and added
//to the destination set. On success one is returned, even if the element was already present in
//the destination set.
//
//An error is raised if the source or destination keys contain a non Set value.
//
//Time complexity O(1)
//param srckey
//param dstkey
//param member
//return Integer reply, specifically: 1 if the element was moved 0 if the element was not found
//        on the first set and no operation was performed
func (r *Redis) Smove(srckey, dstkey, member string) (int64, error) {
	err := r.client.smove(srckey, dstkey, member)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Creates a union or intersection of N sorted sets given by keys k1 through kN, and stores it at
//dstkey. It is mandatory to provide the number of input keys N, before passing the input keys
//and the other (optional) arguments.
//
//As the terms imply, the {@link #zinterstore(String, String...) ZINTERSTORE} command requires an
//element to be present in each of the given inputs to be inserted in the result. The
//{@link #zunionstore(String, String...) ZUNIONSTORE} command inserts all elements across all
//inputs.
//
//Using the WEIGHTS option, it is possible to add weight to each input sorted set. This means
//that the score of each element in the sorted set is first multiplied by this weight before
//being passed to the aggregation. When this option is not given, all weights default to 1.
//
//With the AGGREGATE option, it's possible to specify how the results of the union or
//intersection are aggregated. This option defaults to SUM, where the score of an element is
//summed across the inputs where it exists. When this option is set to be either MIN or MAX, the
//resulting set will contain the minimum or maximum score of an element across the inputs where
//it exists.
//
//<b>Time complexity:</b> O(N) + O(M log(M)) with N being the sum of the sizes of the input
//sorted sets, and M being the number of elements in the resulting sorted set
//@see #zunionstore(String, String...)
//@see #zunionstore(String, ZParams, String...)
//@see #zinterstore(String, String...)
//@see #zinterstore(String, ZParams, String...)
//param dstkey
//param sets
//return Integer reply, specifically the number of elements in the sorted set at dstkey
func (r *Redis) Zunionstore(dstkey string, sets ...string) (int64, error) {
	err := r.client.zunionstore(dstkey, sets...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Creates a union or intersection of N sorted sets given by keys k1 through kN, and stores it at
//dstkey. It is mandatory to provide the number of input keys N, before passing the input keys
//and the other (optional) arguments.
//
//As the terms imply, the {@link #zinterstore(String, String...) ZINTERSTORE} command requires an
//element to be present in each of the given inputs to be inserted in the result. The
//{@link #zunionstore(String, String...) ZUNIONSTORE} command inserts all elements across all
//inputs.
//
//Using the WEIGHTS option, it is possible to add weight to each input sorted set. This means
//that the score of each element in the sorted set is first multiplied by this weight before
//being passed to the aggregation. When this option is not given, all weights default to 1.
//
//With the AGGREGATE option, it's possible to specify how the results of the union or
//intersection are aggregated. This option defaults to SUM, where the score of an element is
//summed across the inputs where it exists. When this option is set to be either MIN or MAX, the
//resulting set will contain the minimum or maximum score of an element across the inputs where
//it exists.
//
//<b>Time complexity:</b> O(N) + O(M log(M)) with N being the sum of the sizes of the input
//sorted sets, and M being the number of elements in the resulting sorted set
//@see #zunionstore(String, String...)
//@see #zunionstore(String, ZParams, String...)
//@see #zinterstore(String, String...)
//@see #zinterstore(String, ZParams, String...)
//param dstkey
//param sets
//return Integer reply, specifically the number of elements in the sorted set at dstkey
func (r *Redis) Zinterstore(dstkey string, sets ...string) (int64, error) {
	err := r.client.zinterstore(dstkey, sets...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//BlpopTimout ...
func (r *Redis) BlpopTimout(timeout int, keys ...string) ([]string, error) {
	err := r.client.blpopTimout(timeout, keys...)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//BrpopTimout ...
func (r *Redis) BrpopTimout(timeout int, keys ...string) ([]string, error) {
	err := r.client.brpopTimout(timeout, keys...)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//BLPOP (and BRPOP) is a blocking list pop primitive. You can see this commands as blocking
//versions of LPOP and RPOP able to block if the specified keys don't exist or contain empty
//lists.
//
//The following is a description of the exact semantic. We describe BLPOP but the two commands
//are identical, the only difference is that BLPOP pops the element from the left (head) of the
//list, and BRPOP pops from the right (tail).
//
//<b>Non blocking behavior</b>
//
//When BLPOP is called, if at least one of the specified keys contain a non empty list, an
//element is popped from the head of the list and returned to the caller together with the name
//of the key (BLPOP returns a two elements array, the first element is the key, the second the
//popped value).
//
//Keys are scanned from left to right, so for instance if you issue BLPOP list1 list2 list3 0
//against a dataset where list1 does not exist but list2 and list3 contain non empty lists, BLPOP
//guarantees to return an element from the list stored at list2 (since it is the first non empty
//list starting from the left).
//
//<b>Blocking behavior</b>
//
//If none of the specified keys exist or contain non empty lists, BLPOP blocks until some other
//client performs a LPUSH or an RPUSH operation against one of the lists.
//
//Once new data is present on one of the lists, the client finally returns with the name of the
//key unblocking it and the popped value.
//
//When blocking, if a non-zero timeout is specified, the client will unblock returning a nil
//special value if the specified amount of seconds passed without a push operation against at
//least one of the specified keys.
//
//The timeout argument is interpreted as an integer value. A timeout of zero means instead to
//block forever.
//
//<b>Multiple clients blocking for the same keys</b>
//
//Multiple clients can block for the same key. They are put into a queue, so the first to be
//served will be the one that started to wait earlier, in a first-blpopping first-served fashion.
//
//<b>blocking POP inside a MULTI/EXEC transaction</b>
//
//BLPOP and BRPOP can be used with pipelining (sending multiple commands and reading the replies
//in batch), but it does not make sense to use BLPOP or BRPOP inside a MULTI/EXEC block (a Redis
//transaction).
//
//The behavior of BLPOP inside MULTI/EXEC when the list is empty is to return a multi-bulk nil
//reply, exactly what happens when the timeout is reached. If you like science fiction, think at
//it like if inside MULTI/EXEC the time will flow at infinite speed :)
//
//Time complexity: O(1)
//@see #brpop(int, String...)
//param timeout
//param keys
//return BLPOP returns a two-elements array via a multi bulk reply in order to return both the
//        unblocking key and the popped value.
//
//        When a non-zero timeout is specified, and the BLPOP operation timed out, the return
//        value is a nil multi bulk reply. Most client values will return false or nil
//        accordingly to the programming language used.
func (r *Redis) Blpop(args ...string) ([]string, error) {
	err := r.client.blpop(args)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//Brpop ...
func (r *Redis) Brpop(args ...string) ([]string, error) {
	err := r.client.brpop(args)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//SortMulti ...
func (r *Redis) SortMulti(key, dstkey string, sortingParameters ...SortingParams) (int64, error) {
	err := r.client.sortMulti(key, dstkey, sortingParameters...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Unwatch ...
func (r *Redis) Unwatch() (string, error) {
	err := r.client.unwatch()
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//ZinterstoreWithParams ...
func (r *Redis) ZinterstoreWithParams(dstkey string, params ZParams, sets ...string) (int64, error) {
	err := r.client.zinterstoreWithParams(dstkey, params, sets...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//ZunionstoreWithParams ...
func (r *Redis) ZunionstoreWithParams(dstkey string, params ZParams, sets ...string) (int64, error) {
	err := r.client.zunionstoreWithParams(dstkey, params, sets...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Brpoplpush ...
func (r *Redis) Brpoplpush(source, destination string, timeout int) (string, error) {
	err := r.client.brpoplpush(source, destination, timeout)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Publish ...
func (r *Redis) Publish(channel, message string) (int64, error) {
	err := r.client.publish(channel, message)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Subscribe ...
func (r *Redis) Subscribe(redisPubSub *RedisPubSub, channels ...string) error {
	err := r.client.connection.setTimeoutInfinite()
	defer r.client.connection.rollbackTimeout()
	if err != nil {
		return err
	}
	err = redisPubSub.proceed(r, channels...)
	if err != nil {
		return err
	}
	return nil
}

//Psubscribe ...
func (r *Redis) Psubscribe(redisPubSub *RedisPubSub, patterns ...string) error {
	err := r.client.connection.setTimeoutInfinite()
	defer r.client.connection.rollbackTimeout()
	if err != nil {
		return err
	}
	err = redisPubSub.proceed(r, patterns...)
	if err != nil {
		return err
	}
	return nil
}

//RandomKey ...
func (r *Redis) RandomKey() (string, error) {
	err := r.client.randomKey()
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Bitop ...
func (r *Redis) Bitop(op BitOP, destKey string, srcKeys ...string) (int64, error) {
	err := r.client.bitop(op, destKey, srcKeys...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Scan ...
func (r *Redis) Scan(cursor string, params ...ScanParams) (*ScanResult, error) {
	err := r.client.scan(cursor, params...)
	if err != nil {
		return nil, err
	}
	return ObjectArrToScanResultReply(r.client.getObjectMultiBulkReply())
}

//Pfmerge ...
func (r *Redis) Pfmerge(destkey string, sourcekeys ...string) (string, error) {
	err := r.client.pfmerge(destkey, sourcekeys...)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

///Pfcount ...
func (r *Redis) Pfcount(keys ...string) (int64, error) {
	err := r.client.pfcount(keys...)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//</editor-fold>

//<editor-fold desc="advancedcommands">

//ConfigGet ...
func (r *Redis) ConfigGet(pattern string) ([]string, error) {
	err := r.client.configGet(pattern)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//ConfigSet ...
func (r *Redis) ConfigSet(parameter, value string) (string, error) {
	err := r.client.configSet(parameter, value)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//SlowlogReset ...
func (r *Redis) SlowlogReset() (string, error) {
	err := r.client.slowlogReset()
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//SlowlogLen ...
func (r *Redis) SlowlogLen() (int64, error) {
	err := r.client.slowlogLen()
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//SlowlogGet ...
func (r *Redis) SlowlogGet(entries ...int64) ([]Slowlog, error) {
	err := r.client.slowlogGet(entries...)
	if err != nil {
		return nil, err
	}
	reply, err := r.client.getObjectMultiBulkReply()
	result := make([]Slowlog, 0)
	for _, re := range reply {
		item := re.([]interface{})
		args := make([]string, 0)
		for _, a := range item[3].([][]byte) {
			args = append(args, string(a))
		}
		result = append(result, Slowlog{
			id:            item[0].(int64),
			timeStamp:     item[1].(int64),
			executionTime: item[2].(int64),
			args:          args,
		})
	}
	return result, err
}

//objectRefcount ...
func (r *Redis) ObjectRefcount(str string) (int64, error) {
	err := r.client.objectRefcount(str)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//ObjectEncoding ...
func (r *Redis) ObjectEncoding(str string) (string, error) {
	err := r.client.objectEncoding(str)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//ObjectIdletime ...
func (r *Redis) ObjectIdletime(str string) (int64, error) {
	err := r.client.objectIdletime(str)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//</editor-fold>

//<editor-fold desc="scriptcommands">

//Eval evaluate scripts using the Lua interpreter built into Redis
func (r *Redis) Eval(script string, keyCount int, params ...string) (interface{}, error) {
	err := r.client.connection.setTimeoutInfinite()
	defer r.client.connection.rollbackTimeout()
	if err != nil {
		return nil, err
	}
	err = r.client.eval(script, keyCount, params...)
	if err != nil {
		return nil, err
	}
	return ObjectToEvalResult(r.client.getOne())
}

//EvalByKeyArgs evaluate scripts using the Lua interpreter built into Redis
func (r *Redis) EvalByKeyArgs(script string, keys []string, args []string) (interface{}, error) {
	err := r.client.connection.setTimeoutInfinite()
	defer r.client.connection.rollbackTimeout()
	if err != nil {
		return nil, err
	}
	params := make([]string, 0)
	params = append(params, keys...)
	params = append(params, args...)
	err = r.client.eval(script, len(keys), params...)
	if err != nil {
		return nil, err
	}
	return ObjectToEvalResult(r.client.getOne())
}

//Evalsha Evaluates a script cached on the server side by its SHA1 digest.
// Scripts are cached on the server side using the SCRIPT LOAD command.
// The command is otherwise identical to EVAL.
func (r *Redis) Evalsha(sha1 string, keyCount int, params ...string) (interface{}, error) {
	err := r.client.evalsha(sha1, keyCount, params...)
	if err != nil {
		return 0, err
	}
	return ObjectToEvalResult(r.client.getOne())
}

//ScriptExists ...
func (r *Redis) ScriptExists(sha1 ...string) ([]bool, error) {
	err := r.client.scriptExists(sha1...)
	if err != nil {
		return nil, err
	}
	reply, err := r.client.getIntegerMultiBulkReply()
	if err != nil {
		return nil, err
	}
	arr := make([]bool, 0)
	for _, re := range reply {
		arr = append(arr, re == 1)
	}
	return arr, nil
}

//ScriptLoad ...
func (r *Redis) ScriptLoad(script string) (string, error) {
	err := r.client.scriptLoad(script)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//</editor-fold>

//<editor-fold desc="basiccommands">

//Quit Ask the server to close the connection.
// The connection is closed as soon as all pending replies have been written to the client.
//
//Return value
//Simple string reply: always OK.
func (r *Redis) Quit() (string, error) {
	err := r.client.quit()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//Ping ...
func (r *Redis) Ping() (string, error) {
	err := r.client.ping()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//Select ...
func (r *Redis) Select(index int) (string, error) {
	err := r.client.selectDb(index)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//FlushDB ...
func (r *Redis) FlushDB() (string, error) {
	err := r.client.flushDB()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//DbSize ...
func (r *Redis) DbSize() (int64, error) {
	err := r.client.dbSize()
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//FlushAll ...
func (r *Redis) FlushAll() (string, error) {
	err := r.client.flushAll()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//Auth ...
func (r *Redis) Auth(password string) (string, error) {
	err := r.client.auth(password)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//Save ...
func (r *Redis) Save() (string, error) {
	err := r.client.save()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//Bgsave ...
func (r *Redis) Bgsave() (string, error) {
	err := r.client.bgsave()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//Bgrewriteaof ...
func (r *Redis) Bgrewriteaof() (string, error) {
	err := r.client.bgrewriteaof()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//Lastsave ...
func (r *Redis) Lastsave() (int64, error) {
	err := r.client.lastsave()
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//Shutdown ...
func (r *Redis) Shutdown() (string, error) {
	err := r.client.shutdown()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//Info ...
func (r *Redis) Info(section ...string) (string, error) {
	err := r.client.info(section...)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//Slaveof ...
func (r *Redis) Slaveof(host string, port int) (string, error) {
	err := r.client.slaveof(host, port)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//SlaveofNoOne ...
func (r *Redis) SlaveofNoOne() (string, error) {
	err := r.client.slaveofNoOne()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//GetDB ...
func (r *Redis) GetDB() int {
	return r.client.Db
}

//Debug ...
func (r *Redis) Debug(params DebugParams) (string, error) {
	err := r.client.debug(params)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ConfigResetStat ...
func (r *Redis) ConfigResetStat() (string, error) {
	err := r.client.configResetStat()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//WaitReplicas ...
func (r *Redis) WaitReplicas(replicas int, timeout int64) (int64, error) {
	err := r.client.waitReplicas(replicas, timeout)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//</editor-fold>

//<editor-fold desc="clustercommands">

//ClusterNodes Each node in a Redis Cluster has its view of the current cluster configuration,
// given by the set of known nodes, the state of the connection we have with such nodes,
// their flags, properties and assigned slots, and so forth.
func (r *Redis) ClusterNodes() (string, error) {
	err := r.client.clusterNodes()
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//ClusterMeet ...
func (r *Redis) ClusterMeet(ip string, port int) (string, error) {
	err := r.client.clusterMeet(ip, port)
	if err != nil {
		return "", err
	}
	return r.client.getBulkReply()
}

//ClusterAddSlots ...
func (r *Redis) ClusterAddSlots(slots ...int) (string, error) {
	err := r.client.clusterAddSlots(slots...)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ClusterDelSlots ...
func (r *Redis) ClusterDelSlots(slots ...int) (string, error) {
	err := r.client.clusterDelSlots(slots...)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ClusterInfo ...
func (r *Redis) ClusterInfo() (string, error) {
	err := r.client.clusterInfo()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ClusterGetKeysInSlot ...
func (r *Redis) ClusterGetKeysInSlot(slot int, count int) ([]string, error) {
	err := r.client.clusterGetKeysInSlot(slot, count)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//ClusterSetSlotNode ...
func (r *Redis) ClusterSetSlotNode(slot int, nodeId string) (string, error) {
	err := r.client.clusterSetSlotNode(slot, nodeId)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ClusterSetSlotMigrating ...
func (r *Redis) ClusterSetSlotMigrating(slot int, nodeId string) (string, error) {
	err := r.client.clusterSetSlotMigrating(slot, nodeId)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ClusterSetSlotImporting ...
func (r *Redis) ClusterSetSlotImporting(slot int, nodeId string) (string, error) {
	err := r.client.clusterSetSlotImporting(slot, nodeId)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ClusterSetSlotStable ...
func (r *Redis) ClusterSetSlotStable(slot int) (string, error) {
	err := r.client.clusterSetSlotStable(slot)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ClusterForget ...
func (r *Redis) ClusterForget(nodeId string) (string, error) {
	err := r.client.clusterForget(nodeId)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ClusterFlushSlots ...
func (r *Redis) ClusterFlushSlots() (string, error) {
	err := r.client.clusterFlushSlots()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ClusterKeySlot ...
func (r *Redis) ClusterKeySlot(key string) (int64, error) {
	err := r.client.clusterKeySlot(key)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

// ClusterCountKeysInSlot ...
func (r *Redis) ClusterCountKeysInSlot(slot int) (int64, error) {
	err := r.client.clusterCountKeysInSlot(slot)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//ClusterSaveConfig ...
func (r *Redis) ClusterSaveConfig() (string, error) {
	err := r.client.clusterSaveConfig()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ClusterReplicate ...
func (r *Redis) ClusterReplicate(nodeId string) (string, error) {
	err := r.client.clusterReplicate(nodeId)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ClusterSlaves ...
func (r *Redis) ClusterSlaves(nodeId string) ([]string, error) {
	err := r.client.clusterSlaves(nodeId)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

//ClusterFailover This command, that can only be sent to a Redis Cluster replica node,
// forces the replica to start a manual failover of its master instance.
func (r *Redis) ClusterFailover() (string, error) {
	err := r.client.clusterFailover()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//ClusterSlots ...
func (r *Redis) ClusterSlots() ([]interface{}, error) {
	err := r.client.clusterSlots()
	if err != nil {
		return nil, err
	}
	return r.client.getObjectMultiBulkReply()
}

//ClusterReset ...
func (r *Redis) ClusterReset(resetType Reset) (string, error) {
	err := r.client.clusterReset(resetType)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//Enables read queries for a connection to a Redis Cluster replica node.
//
//Normally replica nodes will redirect clients to the authoritative master for the hash slot involved in a given command, however clients can use replicas in order to scale reads using the READONLY command.
//
//READONLY tells a Redis Cluster replica node that the client is willing to read possibly stale data and is not interested in running write queries.
//
//When the connection is in readonly mode, the cluster will send a redirection to the client only if the operation involves keys not served by the replica's master node. This may happen because:
//
//The client sent a command about hash slots never served by the master of this replica.
//The cluster was reconfigured (for example resharded) and the replica is no longer able to serve commands for a given hash slot.
//Return value
//Simple string reply
func (r *Redis) Readonly() (string, error) {
	err := r.client.readonly()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//</editor-fold>

//<editor-fold desc="sentinelcommands">

//<pre>
//redis 127.0.0.1:26381&gt; sentinel masters
//1)  1) "name"
//    2) "mymaster"
//    3) "ip"
//    4) "127.0.0.1"
//    5) "port"
//    6) "6379"
//    7) "runid"
//    8) "93d4d4e6e9c06d0eea36e27f31924ac26576081d"
//    9) "flags"
//   10) "master"
//   11) "pending-commands"
//   12) "0"
//   13) "last-ok-ping-reply"
//   14) "423"
//   15) "last-ping-reply"
//   16) "423"
//   17) "info-refresh"
//   18) "6107"
//   19) "num-slaves"
//   20) "1"
//   21) "num-other-sentinels"
//   22) "2"
//   23) "quorum"
//   24) "2"
//</pre>
//return
func (r *Redis) SentinelMasters() ([]map[string]string, error) {
	err := r.client.sentinelMasters()
	if err != nil {
		return nil, err
	}
	return ObjectArrToMapArrayReply(r.client.getObjectMultiBulkReply())
}

//<pre>
//redis 127.0.0.1:26381&gt; sentinel get-master-addr-by-name mymaster
//1) "127.0.0.1"
//2) "6379"
//</pre>
//param masterName
//return two elements list of strings : host and port.
func (r *Redis) SentinelGetMasterAddrByName(masterName string) ([]string, error) {
	err := r.client.sentinelGetMasterAddrByName(masterName)
	if err != nil {
		return nil, err
	}
	reply, err := r.client.getObjectMultiBulkReply()
	if err != nil {
		return nil, err
	}
	addrs := make([]string, 0)
	for _, re := range reply {
		if re == nil {
			addrs = append(addrs, "")
		} else {
			addrs = append(addrs, string(re.([]byte)))
		}
	}
	return addrs, err
}

//<pre>
//redis 127.0.0.1:26381&gt; sentinel reset mymaster
//(integer) 1
//</pre>
//param pattern
//return
func (r *Redis) SentinelReset(pattern string) (int64, error) {
	err := r.client.sentinelReset(pattern)
	if err != nil {
		return 0, err
	}
	return r.client.getIntegerReply()
}

//<pre>
//redis 127.0.0.1:26381&gt; sentinel slaves mymaster
//1)  1) "name"
//    2) "127.0.0.1:6380"
//    3) "ip"
//    4) "127.0.0.1"
//    5) "port"
//    6) "6380"
//    7) "runid"
//    8) "d7f6c0ca7572df9d2f33713df0dbf8c72da7c039"
//    9) "flags"
//   10) "slave"
//   11) "pending-commands"
//   12) "0"
//   13) "last-ok-ping-reply"
//   14) "47"
//   15) "last-ping-reply"
//   16) "47"
//   17) "info-refresh"
//   18) "657"
//   19) "master-link-down-time"
//   20) "0"
//   21) "master-link-status"
//   22) "ok"
//   23) "master-host"
//   24) "localhost"
//   25) "master-port"
//   26) "6379"
//   27) "slave-priority"
//   28) "100"
//</pre>
//param masterName
//return
func (r *Redis) SentinelSlaves(masterName string) ([]map[string]string, error) {
	err := r.client.sentinelSlaves(masterName)
	if err != nil {
		return nil, err
	}
	return ObjectArrToMapArrayReply(r.client.getObjectMultiBulkReply())
}

//SentinelFailover ...
func (r *Redis) SentinelFailover(masterName string) (string, error) {
	err := r.client.sentinelFailover(masterName)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

// SentinelMonitor ...
func (r *Redis) SentinelMonitor(masterName, ip string, port, quorum int) (string, error) {
	err := r.client.sentinelMonitor(masterName, ip, port, quorum)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

// SentinelRemove ...
func (r *Redis) SentinelRemove(masterName string) (string, error) {
	err := r.client.sentinelRemove(masterName)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

// SentinelSet ...
func (r *Redis) SentinelSet(masterName string, parameterMap map[string]string) (string, error) {
	err := r.client.sentinelSet(masterName, parameterMap)
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

//</editor-fold>

//<editor-fold desc="other commands">

// PubsubChannels ...
func (r *Redis) PubsubChannels(pattern string) ([]string, error) {
	err := r.client.pubsubChannels(pattern)
	if err != nil {
		return nil, err
	}
	return r.client.getMultiBulkReply()
}

// Asking ...
func (r *Redis) Asking() (string, error) {
	err := r.client.asking()
	if err != nil {
		return "", err
	}
	return r.client.getStatusCodeReply()
}

// get transaction of redis client ,when use transaction mode, you need to invoke this first
func (r *Redis) Multi() (*transaction, error) {
	err := r.client.multi()
	if err != nil {
		return nil, err
	}
	return newTransaction(r.client), nil
}

// get pipeline of redis client ,when use pipeline mode, you need to invoke this first
func (r *Redis) Pipelined() *pipeline {
	return newPipeline(r.client)
}

//</editor-fold>
