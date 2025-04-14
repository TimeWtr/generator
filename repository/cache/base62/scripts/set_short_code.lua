-- 插入一条短码到短码池
-- 1. 插入一条短码
-- 2. 短码数量计数+1

local poolKey = KEYS[1]
local countKey = KEYS[2]
local code = ARGV[1]

local res = redis.call("LPUSH", poolKey, code)
if res > 0 then
	redis.call("INCR", countKey)
	return 0
else
	return -1
end