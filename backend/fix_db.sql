-- 删除email字段为空的记录
DELETE FROM users WHERE email = '' OR email IS NULL;

-- 如果仍有重复的email，保留id最小的记录
DELETE u1 FROM users u1
INNER JOIN users u2 
WHERE u1.email = u2.email AND u1.id > u2.id;

-- 清理username重复
DELETE u1 FROM users u1
INNER JOIN users u2 
WHERE u1.username = u2.username AND u1.id > u2.id;