-- 流程
-- 1. 获取到一个可用的短码
-- 2. 更新可用数量计数

-- 缓存池的Key
local poolKey = KEYS[1]
-- 缓存数量的Key
local countKey = KEYS[2]

local function isEmpty(s)
    return s == "" or s == nil
end

local val = redis.call("RPOP", poolKey)
if isEmpty(val) then
   -- 没有获取到可用的短码
   return ""
else
	redis.call("INCRBY", countKey, -1)
	return val
end