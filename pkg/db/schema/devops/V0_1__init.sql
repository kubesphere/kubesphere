CREATE TABLE project (
  `project_id`  VARCHAR(50)        NOT NULL,
  `name`        VARCHAR(50)        NOT NULL,
  `description` TEXT               NOT NULL,
  `creator`     VARCHAR(50)        NOT NULL,
  `create_time` TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status`      VARCHAR(50)        NOT NULL,
  `visibility`  VARCHAR(50)        NOT NULL,
  `extra`       TEXT               NOT NULL,
  PRIMARY KEY (`project_id`)
);



CREATE TABLE `project_membership` (
  `username`   VARCHAR(50) NOT NULL,
  `project_id` VARCHAR(50) NOT NULL,
  `role`       VARCHAR(50) NOT NULL,
  `status`     VARCHAR(50) NOT NULL,
  `grant_by`   VARCHAR(50) NOT NULL,
  PRIMARY KEY (`username`, `project_id`)
);

