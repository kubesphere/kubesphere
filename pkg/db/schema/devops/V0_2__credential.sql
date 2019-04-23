CREATE TABLE `project_credential` (
  `project_id`    VARCHAR(50)  NOT NULL,
  `credential_id` VARCHAR(255) NOT NULL,
  `domain`        VARCHAR(255) NOT NULL,
  `creator`       VARCHAR(50)  NOT NULL,
  `create_time`   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`project_id`, `credential_id`, `domain`)
);
