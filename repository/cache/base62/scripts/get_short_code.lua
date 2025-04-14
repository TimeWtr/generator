-- 从短码池中获取一个可用的短码
-- 1. 从池中获取一个短码
-- 2. 短码数量计数-1

local poolKey = KEYS[1]
local countKey = KEYS[2]

local function isEmpty(s) 
	return s == "" or s == nil
end

lcoal val = redis.call("RPOP", poolKey)
if isEmpty(val) then
	redis.call("INCRBY", countKey, -1)
	return val
else
	return ""
end