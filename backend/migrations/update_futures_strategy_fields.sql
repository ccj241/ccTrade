-- 更新期货策略表结构
-- 添加新字段
ALTER TABLE `futures_strategies` 
ADD COLUMN `margin_amount` DECIMAL(20,8) DEFAULT 0 COMMENT '保证金本值（USDT）' AFTER `side`,
ADD COLUMN `float_basis_points` INT DEFAULT 0 COMMENT '万分比浮动（正数）' AFTER `price`,
ADD COLUMN `take_profit_bp` INT DEFAULT 0 COMMENT '止盈万分比' AFTER `float_basis_points`,
ADD COLUMN `stop_loss_bp` INT DEFAULT 0 COMMENT '止损万分比' AFTER `take_profit_bp`;

-- 更新默认杠杆倍数为8
ALTER TABLE `futures_strategies` 
MODIFY COLUMN `leverage` INT DEFAULT 8 COMMENT '杠杆倍数，默认8倍';

-- 更新默认保证金模式为逐仓
ALTER TABLE `futures_strategies` 
MODIFY COLUMN `margin_type` VARCHAR(20) DEFAULT 'isolated' COMMENT '保证金模式，默认逐仓';

-- 移除不再使用的字段（可选，建议先保留以备数据迁移）
-- ALTER TABLE `futures_strategies` 
-- DROP COLUMN `position_side`,
-- DROP COLUMN `quantity`,
-- DROP COLUMN `trigger_price`,
-- DROP COLUMN `stop_price`,
-- DROP COLUMN `take_profit`,
-- DROP COLUMN `stop_loss`;

-- 数据迁移（将旧字段数据转换为新字段）
UPDATE `futures_strategies` 
SET 
    `margin_amount` = CASE 
        WHEN `quantity` > 0 AND `price` > 0 AND `leverage` > 0 
        THEN (`quantity` * `price`) / `leverage`
        ELSE 100 -- 默认100 USDT
    END,
    `take_profit_bp` = CASE 
        WHEN `take_profit` > 0 AND `price` > 0 
        THEN ROUND(ABS(`take_profit` - `price`) / `price` * 10000)
        ELSE 0
    END,
    `stop_loss_bp` = CASE 
        WHEN `stop_loss` > 0 AND `price` > 0 
        THEN ROUND(ABS(`stop_loss` - `price`) / `price` * 10000)
        ELSE 0
    END
WHERE 1=1;

-- 添加索引优化查询性能
CREATE INDEX `idx_futures_strategies_margin_amount` ON `futures_strategies` (`margin_amount`);
CREATE INDEX `idx_futures_strategies_leverage` ON `futures_strategies` (`leverage`);