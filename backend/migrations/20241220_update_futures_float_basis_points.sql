-- 修改futures_strategies表的float_basis_points字段类型
ALTER TABLE futures_strategies 
MODIFY COLUMN float_basis_points DECIMAL(10,4) DEFAULT 0.1 COMMENT '首单万分比浮动（支持小数）';

-- 更新现有数据，将整数值转换为小数（如果需要）
UPDATE futures_strategies 
SET float_basis_points = float_basis_points / 1.0 
WHERE float_basis_points IS NOT NULL;