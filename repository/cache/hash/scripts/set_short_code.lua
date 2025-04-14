-- 插入单条短码
-- 1. 插入一条数据
-- 2. 修改数据
-- 3. 插入到BloomFilter

-- 获取到操作的key
local poolKey = KEYS[1]
local countKey = KEYS[2]
local bfKey = KEYS[3]

-- 获取新的短码
local shortCode = ARGV[1]

local res = redis.call("LPUSH", poolKey, shortCode)
if res > 0 then
	-- 写入成功
	redis.call("INCR", countKey)
	redis.call("BF.ADD", bfKey, shortCode)

	return 0
else
	-- 写入失败
	return -1
end
