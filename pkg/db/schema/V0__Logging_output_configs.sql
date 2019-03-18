CREATE TABLE output_db_bindings
	(
		id INT NOT NULL AUTO_INCREMENT,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		parameters TEXT NOT NULL,
		internal BOOLEAN,
		enable BOOLEAN NOT NULL,
		updatetime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE (id)
	);

INSERT INTO output_db_bindings (type, name, parameters, enable) VALUES ('fluentbit_output', 'fluentbit-output', '[{"name": "Name", "value": "es"}, {"name": "Match", "value": "kube.*"}, {"name": "Host", "value": "elasticsearch-logging-data.kubesphere-logging-system.svc"}, {"name": "Port", "value": "9200"}, {"name": "Logstash_Format", "value": "On"}, {"name": "Replace_Dots", "value": "on"}, {"name": "Retry_Limit", "value": "False"}, {"name": "Type", "value": "flb_type"}, {"name": "Time_Key", "value": "@timestamp"}, {"name": "Logstash_Prefix", "value": "logstash"} ]', '1');