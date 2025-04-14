-- 批量写入短码到短码池
-- 1. 写入短码到短码池
-- 2. 更新短码计数
-- 3. 短码写入过滤器

local poolKey = KEYS[1]
local countKey = KEYS[2]
local bfKey = KEYS[3]
local codes = ARGV[1]
local batchCount = tonumber(ARGV[2])

local val = redis.call("LPUSH", poolKey, unpack(codes))
if val > 0 then
    redis.call("INCRBY", countKey, batchCount)
    redis.call("BF.MADD", bfKey, unpack(codes))
    return 0
else
	return -1
end