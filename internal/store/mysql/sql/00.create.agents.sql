CREATE TABLE IF NOT EXISTS
    Agents_V1(
        id TEXT(36) NOT NULL PRIMARY KEY,
        name TEXT(255) NOT NULL,
        identity TEXT
    );

-- CREATE UNIQUE INDEX IF NOT EXISTS agents_id_v1 ON Agents_V1 (id);

IF (SELECT 1        
    FROM `INFORMATION_SCHEMA`.`STATISTICS`
    WHERE `TABLE_NAME` = 'Agents_V1'
    AND `INDEX_NAME` = 'agents_id_v1') IS NULL THEN

    ALTER TABLE `Agents_V1` ADD INDEX `agents_id_v1` (`id`);

END IF;


-- DELIMITER $$

-- CREATE PROCEDURE create_unique_index_agents_id_v1()
-- BEGIN
--   DECLARE index_exists INT DEFAULT 0;

--   SELECT COUNT(1) INTO index_exists
--   FROM information_schema.statistics
--   WHERE table_name = 'Agents_V1'
--     AND index_name = 'agents_id_v1';

--   IF index_exists = 0 THEN
--     CREATE UNIQUE INDEX agents_id_v1 ON Agents_V1 (id);
--   END IF;
-- END$$

-- -- DELIMITER ;

-- CALL create_unique_index_agents_id_v1