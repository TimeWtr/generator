-- 批量插入短码到短码池
-- 1. 批量插入数据
-- 2. 短码数量计数增加

local poolKey = KEYS[1]
local countKey = KEYS[2]
local codes = ARGV[1]
local codesCount = tonumber(ARGV[2])

local res = redis.call("LPUSH", poolKey, unpack(codes))
if res > 0 then
	redis.call("INCRBY", countKey, codesCount)
	return 0
else
	return -1
end
