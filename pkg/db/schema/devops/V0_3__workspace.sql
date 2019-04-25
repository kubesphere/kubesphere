CREATE TABLE IF NOT EXISTS `kubesphere`.`workspace_dp_bindings`  (
  `workspace` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `dev_ops_project` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`workspace`,`dev_ops_project`)
) ENGINE=InnoDB;

ALTER TABLE kubesphere.workspace_dp_bindings
CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

ALTER TABLE devops.project
ADD COLUMN workspace VARCHAR(255) NOT NULL DEFAULT '';

UPDATE devops.project t1
INNER JOIN kubesphere.workspace_dp_bindings t2 ON t1.project_id= t2.dev_ops_project
SET t1.workspace=t2.workspace;
