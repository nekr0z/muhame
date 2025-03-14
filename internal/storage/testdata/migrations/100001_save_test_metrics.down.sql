BEGIN;

DELETE FROM counters WHERE name=='test_counter';

DELETE FROM gauges WHERE name=='test_gauge';

COMMIT;