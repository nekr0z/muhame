BEGIN;

INSERT INTO counters(name, value) VALUES ('test_counter', 11);

INSERT INTO gauges(name, value) VALUES ('test_gauge', 1.2);

COMMIT;