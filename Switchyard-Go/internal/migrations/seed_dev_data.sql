-- Dev seed data — drivers, HOS windows, equipment
-- Run from psql: \i internal/migrations/seed_dev_data.sql
-- Auth0 IDs are placeholders; replace with real user IDs before any Auth0-gated flow.

-- ── Drivers ──────────────────────────────────────────────────────────────────
INSERT INTO driver (id, name, home_warehouse_id, auth0_user_id, license_state, is_active) VALUES
  ('a1000000-0000-0000-0000-000000000001', 'Marcus Webb',     'WH001', 'auth0|testdriver001', 'OH', true),
  ('a1000000-0000-0000-0000-000000000002', 'Diane Kowalski',  'WH001', 'auth0|testdriver002', 'OH', true),
  ('a1000000-0000-0000-0000-000000000003', 'Ray Gutierrez',   'WH002', 'auth0|testdriver003', 'IN', true),
  ('a1000000-0000-0000-0000-000000000004', 'Sandra Okafor',   'WH002', 'auth0|testdriver004', 'IN', true),
  ('a1000000-0000-0000-0000-000000000005', 'Tom Brierley',    'WH003', 'auth0|testdriver005', 'IL', true),
  ('a1000000-0000-0000-0000-000000000006', 'Keisha Drummond', 'WH003', 'auth0|testdriver006', 'IL', true),
  ('a1000000-0000-0000-0000-000000000007', 'Pete Halverson',  'WH004', 'auth0|testdriver007', 'MI', true),
  ('a1000000-0000-0000-0000-000000000008', 'Angela Torres',   'WH004', 'auth0|testdriver008', 'MI', true),
  ('a1000000-0000-0000-0000-000000000009', 'James Whitfield', 'WH005', 'auth0|testdriver009', 'WI', true),
  ('a1000000-0000-0000-0000-000000000010', 'Carol Metzger',   'WH005', 'auth0|testdriver010', 'WI', true);

-- ── HOS Windows ──────────────────────────────────────────────────────────────
-- Green  = daily < 8h, weekly < 55h  (ample headroom)
-- Yellow = daily 9–10h or weekly 60–65h  (approaching limit)
-- Window start = beginning of current work week
INSERT INTO hos_window (id, driver_id, window_start, daily_hours_used, weekly_hours_used) VALUES
  -- Green — fresh
  ('b1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', DATE_TRUNC('week', NOW()), 1.5,  12.0),
  ('b1000000-0000-0000-0000-000000000002', 'a1000000-0000-0000-0000-000000000002', DATE_TRUNC('week', NOW()), 0.0,   8.5),
  -- Green — moderate
  ('b1000000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000003', DATE_TRUNC('week', NOW()), 4.5,  32.0),
  ('b1000000-0000-0000-0000-000000000004', 'a1000000-0000-0000-0000-000000000004', DATE_TRUNC('week', NOW()), 5.5,  41.0),
  -- Green — mid-week
  ('b1000000-0000-0000-0000-000000000005', 'a1000000-0000-0000-0000-000000000005', DATE_TRUNC('week', NOW()), 3.0,  28.5),
  ('b1000000-0000-0000-0000-000000000006', 'a1000000-0000-0000-0000-000000000006', DATE_TRUNC('week', NOW()), 6.0,  47.0),
  -- Yellow — approaching daily limit
  ('b1000000-0000-0000-0000-000000000007', 'a1000000-0000-0000-0000-000000000007', DATE_TRUNC('week', NOW()), 9.0,  52.0),
  ('b1000000-0000-0000-0000-000000000008', 'a1000000-0000-0000-0000-000000000008', DATE_TRUNC('week', NOW()), 9.5,  58.0),
  -- Yellow — approaching weekly limit
  ('b1000000-0000-0000-0000-000000000009', 'a1000000-0000-0000-0000-000000000009', DATE_TRUNC('week', NOW()), 7.0,  62.0),
  ('b1000000-0000-0000-0000-000000000010', 'a1000000-0000-0000-0000-000000000010', DATE_TRUNC('week', NOW()), 6.5,  63.5);

-- ── Equipment — Trucks ────────────────────────────────────────────────────────
INSERT INTO equipment (id, unit_id, equipment_type, home_warehouse_id, status) VALUES
  ('c1000000-0000-0000-0000-000000000001', 'TK-101', 'truck', 'WH001', 'available'),
  ('c1000000-0000-0000-0000-000000000002', 'TK-102', 'truck', 'WH001', 'available'),
  ('c1000000-0000-0000-0000-000000000003', 'TK-103', 'truck', 'WH002', 'available'),
  ('c1000000-0000-0000-0000-000000000004', 'TK-104', 'truck', 'WH003', 'available'),
  ('c1000000-0000-0000-0000-000000000005', 'TK-105', 'truck', 'WH004', 'available'),
  ('c1000000-0000-0000-0000-000000000006', 'TK-106', 'truck', 'WH005', 'available');

-- ── Equipment — Trailers ──────────────────────────────────────────────────────
INSERT INTO equipment (id, unit_id, equipment_type, home_warehouse_id, status) VALUES
  ('c1000000-0000-0000-0000-000000000011', 'TR-2001', 'trailer', 'WH001', 'available'),
  ('c1000000-0000-0000-0000-000000000012', 'TR-2002', 'trailer', 'WH001', 'available'),
  ('c1000000-0000-0000-0000-000000000013', 'TR-2003', 'trailer', 'WH002', 'available'),
  ('c1000000-0000-0000-0000-000000000014', 'TR-2004', 'trailer', 'WH002', 'available'),
  ('c1000000-0000-0000-0000-000000000015', 'TR-2005', 'trailer', 'WH003', 'available'),
  ('c1000000-0000-0000-0000-000000000016', 'TR-2006', 'trailer', 'WH004', 'available'),
  ('c1000000-0000-0000-0000-000000000017', 'TR-2007', 'trailer', 'WH005', 'available');

-- ── PlanBOL Records ───────────────────────────────────────────────────────────
-- All timestamps relative to NOW() — board stays fresh on any run date.
-- driver_id is the planned/intended driver; the formal link is driver_bol_assignment.
INSERT INTO plan_bol_record (id, driver_id, originating_wh_id, status, created_at, submitted_at) VALUES
  -- Draft: dispatcher just received invoice ~2h ago, no planning begun
  ('d1000000-0000-0000-0000-000000000001',
   'a1000000-0000-0000-0000-000000000001', 'WH001', 'draft',
   NOW() - INTERVAL '2 hours', NULL),
  -- Pending: claimed for route planning yesterday, work in progress
  ('d1000000-0000-0000-0000-000000000002',
   'a1000000-0000-0000-0000-000000000002', 'WH001', 'plan-progress',
   NOW() - INTERVAL '18 hours', NULL),
  -- Loading: committed to .NET ~4h ago, dock loading trailer now
  ('d1000000-0000-0000-0000-000000000003',
   'a1000000-0000-0000-0000-000000000003', 'WH002', 'loading',
   NOW() - INTERVAL '20 hours', NULL),
  -- Validated: trailer loaded and confirmed this morning, awaiting driver + equipment assignment
  ('d1000000-0000-0000-0000-000000000004',
   'a1000000-0000-0000-0000-000000000004', 'WH002', 'validated',
   NOW() - INTERVAL '22 hours', NULL),
  -- Submitted: departed 4h ago, Tom Brierley currently in transit on stop 2 of 3
  ('d1000000-0000-0000-0000-000000000005',
   'a1000000-0000-0000-0000-000000000005', 'WH003', 'submitted',
   NOW() - INTERVAL '6 hours', NOW() - INTERVAL '4 hours');

-- ── PlanBOL Stops ─────────────────────────────────────────────────────────────
-- Stops only needed for Loading, Validated, and Submitted BOLs (board reads stop count
-- and current stop for those card types). Draft and Pending stops omitted for brevity.
-- delivery_items: {"sku_id": qty} map — warehouse stops are loads, store stops are deliveries.

-- Loading BOL (003) — WH002 pickup + 2 store deliveries
INSERT INTO plan_bol_stop (id, plan_bol_id, sequence, location_id, stop_type, delivery_items, is_processed) VALUES
  ('e1000000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000003', 1, 'WH002',     'warehouse', '{"SKU-GLOVE-L":24,"SKU-VEST-M":12}',     false),
  ('e1000000-0000-0000-0000-000000000002', 'd1000000-0000-0000-0000-000000000003', 2, 'STORE-201', 'store',     '{"SKU-GLOVE-L":12,"SKU-VEST-M":6}',      false),
  ('e1000000-0000-0000-0000-000000000003', 'd1000000-0000-0000-0000-000000000003', 3, 'STORE-202', 'store',     '{"SKU-GLOVE-L":12,"SKU-VEST-M":6}',      false);

-- Validated BOL (004) — WH002 pickup + 2 store deliveries
INSERT INTO plan_bol_stop (id, plan_bol_id, sequence, location_id, stop_type, delivery_items, is_processed) VALUES
  ('e1000000-0000-0000-0000-000000000004', 'd1000000-0000-0000-0000-000000000004', 1, 'WH002',     'warehouse', '{"SKU-BOOT-10":20,"SKU-JACKET-XL":8}',   false),
  ('e1000000-0000-0000-0000-000000000005', 'd1000000-0000-0000-0000-000000000004', 2, 'STORE-203', 'store',     '{"SKU-BOOT-10":10,"SKU-JACKET-XL":4}',   false),
  ('e1000000-0000-0000-0000-000000000006', 'd1000000-0000-0000-0000-000000000004', 3, 'STORE-204', 'store',     '{"SKU-BOOT-10":10,"SKU-JACKET-XL":4}',   false);

-- Submitted BOL (005) — WH003 pickup done, 2 store stops remaining
INSERT INTO plan_bol_stop (id, plan_bol_id, sequence, location_id, stop_type, delivery_items, is_processed, processed_at) VALUES
  ('e1000000-0000-0000-0000-000000000007', 'd1000000-0000-0000-0000-000000000005', 1, 'WH003',     'warehouse', '{"SKU-HARDHAT-M":30,"SKU-GLOVE-M":48}',  true,  NOW() - INTERVAL '5 hours 30 minutes'),
  ('e1000000-0000-0000-0000-000000000008', 'd1000000-0000-0000-0000-000000000005', 2, 'STORE-301', 'store',     '{"SKU-HARDHAT-M":15,"SKU-GLOVE-M":24}',  false, NULL),
  ('e1000000-0000-0000-0000-000000000009', 'd1000000-0000-0000-0000-000000000005', 3, 'STORE-302', 'store',     '{"SKU-HARDHAT-M":15,"SKU-GLOVE-M":24}',  false, NULL);

-- ── Assignment — Tom Brierley, BOL 005, TK-104 ────────────────────────────────
-- departed_at set → board places this card in In Delivery > In Transit.
-- fulfilled_at null → dead-head timer not running yet.
-- estimated_run_hours puts estimated fulfillment ~15 min out — inside
-- DEADHEAD_CUTOFF_MINUTES (30) with no pairing secured, so this card demos
-- the pairing-at-risk warn border (ARCHITECTURE.md §5).
INSERT INTO driver_bol_assignment (id, driver_id, plan_bol_id, equipment_id, assigned_at, departed_at, estimated_run_hours) VALUES
  ('f1000000-0000-0000-0000-000000000001',
   'a1000000-0000-0000-0000-000000000005',
   'd1000000-0000-0000-0000-000000000005',
   'c1000000-0000-0000-0000-000000000004',
   NOW() - INTERVAL '5 hours',
   NOW() - INTERVAL '4 hours',
   4.25);

-- ── BOL 006 — completed run, no dead-head pairing (Empty Return demo) ────────
-- Keisha Drummond completed her WH003 run 20 minutes ago with no pairing
-- secured — demos both the Delivered close-out card and the Empty Return
-- sub-section (ARCHITECTURE.md §3-4).
INSERT INTO plan_bol_record (id, driver_id, originating_wh_id, status, created_at, submitted_at, fulfilled_at) VALUES
  ('d1000000-0000-0000-0000-000000000006',
   'a1000000-0000-0000-0000-000000000006', 'WH003', 'fulfilled',
   NOW() - INTERVAL '6 hours', NOW() - INTERVAL '5 hours', NOW() - INTERVAL '20 minutes');

INSERT INTO plan_bol_stop (id, plan_bol_id, sequence, location_id, stop_type, delivery_items, is_processed, processed_at) VALUES
  ('e1000000-0000-0000-0000-000000000010', 'd1000000-0000-0000-0000-000000000006', 1, 'WH003',     'warehouse', '{"SKU-VEST-M":20}', true, NOW() - INTERVAL '5 hours'),
  ('e1000000-0000-0000-0000-000000000011', 'd1000000-0000-0000-0000-000000000006', 2, 'STORE-303', 'store',     '{"SKU-VEST-M":20}', true, NOW() - INTERVAL '20 minutes');

INSERT INTO driver_bol_assignment (id, driver_id, plan_bol_id, equipment_id, assigned_at, departed_at, fulfilled_at) VALUES
  ('f1000000-0000-0000-0000-000000000002',
   'a1000000-0000-0000-0000-000000000006',
   'd1000000-0000-0000-0000-000000000006',
   'c1000000-0000-0000-0000-000000000015',
   NOW() - INTERVAL '5 hours',
   NOW() - INTERVAL '4 hours 30 minutes',
   NOW() - INTERVAL '20 minutes');

-- Equipment mirrors what AssignmentHandler.Fulfill would have done: no pairing
-- secured, so the trailer stays Assigned under its own empty-return timer.
UPDATE equipment SET status = 'assigned', empty_return_until = NOW() + INTERVAL '1 hour 40 minutes'
  WHERE id = 'c1000000-0000-0000-0000-000000000015';
