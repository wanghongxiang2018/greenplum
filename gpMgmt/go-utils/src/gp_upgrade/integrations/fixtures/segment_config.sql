-- test rig bootstrap:

-- gp_dump -D -t gp_segment_configuration template1

CREATE TABLE gp_segment_configuration (
  dbid             SMALLINT NOT NULL,
  content          SMALLINT NOT NULL,
  role             "CHAR"   NOT NULL,
  preferred_role   "CHAR"   NOT NULL,
  mode             "CHAR"   NOT NULL,
  status           "CHAR"   NOT NULL,
  port             INTEGER  NOT NULL,
  hostname         TEXT,
  address          TEXT,
  datadir          TEXT
);

INSERT INTO gp_segment_configuration (dbid, content, role, preferred_role, mode, status, port, hostname, address, datadir) VALUES (1, -1, 'p', 'p', 's', 'u', 15432, 'office-5-231.pa.pivotal.io', 'office-5-231.pa.pivotal.io', '/Users/pivotal/workspace/gpdb/gpAux/gpdemo/datadirs/qddir/demoDataDir-1');
INSERT INTO gp_segment_configuration (dbid, content, role, preferred_role, mode, status, port, hostname, address, datadir) VALUES (2, 0, 'p', 'p', 's', 'u', 25432, 'office-5-231.pa.pivotal.io', 'office-5-231.pa.pivotal.io', '/Users/pivotal/workspace/gpdb/gpAux/gpdemo/datadirs/dbfast1/demoDataDir0');
INSERT INTO gp_segment_configuration (dbid, content, role, preferred_role, mode, status, port, hostname, address, datadir) VALUES (3, 1, 'p', 'p', 's', 'u', 25433, 'office-5-231.pa.pivotal.io', 'office-5-231.pa.pivotal.io', '/Users/pivotal/workspace/gpdb/gpAux/gpdemo/datadirs/dbfast2/demoDataDir1');
INSERT INTO gp_segment_configuration (dbid, content, role, preferred_role, mode, status, port, hostname, address, datadir) VALUES (4, 2, 'p', 'p', 's', 'u', 25434, 'office-5-231.pa.pivotal.io', 'office-5-231.pa.pivotal.io', '/Users/pivotal/workspace/gpdb/gpAux/gpdemo/datadirs/dbfast3/demoDataDir2');
INSERT INTO gp_segment_configuration (dbid, content, role, preferred_role, mode, status, port, hostname, address, datadir) VALUES (5, 0, 'm', 'm', 's', 'u', 25435, 'office-5-231.pa.pivotal.io', 'office-5-231.pa.pivotal.io', '/Users/pivotal/workspace/gpdb/gpAux/gpdemo/datadirs/dbfast_mirror1/demoDataDir0');
INSERT INTO gp_segment_configuration (dbid, content, role, preferred_role, mode, status, port, hostname, address, datadir) VALUES (6, 1, 'm', 'm', 's', 'u', 25436, 'office-5-231.pa.pivotal.io', 'office-5-231.pa.pivotal.io', '/Users/pivotal/workspace/gpdb/gpAux/gpdemo/datadirs/dbfast_mirror2/demoDataDir1');
INSERT INTO gp_segment_configuration (dbid, content, role, preferred_role, mode, status, port, hostname, address, datadir) VALUES (7, 2, 'm', 'm', 's', 'u', 25437, 'office-5-231.pa.pivotal.io', 'office-5-231.pa.pivotal.io', '/Users/pivotal/workspace/gpdb/gpAux/gpdemo/datadirs/dbfast_mirror3/demoDataDir2');

