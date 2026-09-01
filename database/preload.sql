
INSERT INTO widget_type (widget_type_name) 
VALUES ('BarChart'),('BarLineChart'),('BulletChart'),('AnalogueGauge'),('DigitalGauge'),('LineChart'),('PieChart'),('ScatterChart'),('BarProcess'),('Status'),('Table'),('Alert'),('Text'),('RowChart'),('ScoreCard'),('InputAction'),('Visualization')
ON CONFLICT (widget_type_name) DO NOTHING;

-- Create Menus
INSERT INTO menu (menu_name) 
VALUES ('Dashboard'),('Canvas'),('Canvas Design'),('Canvas Access'),('Scheduler'),('Log Report'),('Notification User'),('Notification Device'),('User'),('Role'),('Device'),('Device Group')
ON CONFLICT (menu_name) DO NOTHING;

-- Create Action
INSERT INTO action (action_name) 
VALUES  ('Display'),('Create'),('Update'),('Delete'),('Import'),('Export'),('Query'),('Command')
ON CONFLICT (action_name) DO NOTHING;

-- Insert Action To Menu
INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu m
CROSS JOIN action a
WHERE m.menu_name = 'Dashboard' 
  AND a.action_name IN ('Display','Query')
ON CONFLICT (menu_id, action_id) DO NOTHING;

INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu m
CROSS JOIN action a
WHERE m.menu_name = 'Device' 
  AND a.action_name IN ('Display', 'Create', 'Update', 'Delete', 'Import','Export','Command')
ON CONFLICT (menu_id, action_id) DO NOTHING;

INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu m
CROSS JOIN action a
WHERE m.menu_name = 'Device Group' 
  AND a.action_name IN ('Display', 'Create', 'Update', 'Delete')
ON CONFLICT (menu_id, action_id) DO NOTHING;

INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu m
CROSS JOIN action a
WHERE m.menu_name = 'User' 
  AND a.action_name IN ('Display', 'Create', 'Update','Delete')
ON CONFLICT (menu_id, action_id) DO NOTHING;

INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu m
CROSS JOIN action a
WHERE m.menu_name = 'Role' 
  AND a.action_name IN ('Display', 'Create', 'Update')
ON CONFLICT (menu_id, action_id) DO NOTHING;

INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu m
CROSS JOIN action a
WHERE m.menu_name = 'Canvas' 
  AND a.action_name IN ('Display', 'Create','Update','Delete')
ON CONFLICT (menu_id, action_id) DO NOTHING;

INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu m
CROSS JOIN action a
WHERE m.menu_name = 'Canvas Design' 
  AND a.action_name IN ('Display', 'Update')
ON CONFLICT (menu_id, action_id) DO NOTHING;

INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu m
CROSS JOIN action a
WHERE m.menu_name = 'Canvas Access' 
  AND a.action_name IN ('Display', 'Update')
ON CONFLICT (menu_id, action_id) DO NOTHING;

INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu m
CROSS JOIN action a
WHERE m.menu_name = 'Scheduler' 
  AND a.action_name IN ('Display','Create', 'Update','Delete')
ON CONFLICT (menu_id, action_id) DO NOTHING;

INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu m
CROSS JOIN action a
WHERE m.menu_name = 'Log Report' 
  AND a.action_name IN ('Display','Export')
ON CONFLICT (menu_id, action_id) DO NOTHING;

INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu m
CROSS JOIN action a
WHERE m.menu_name = 'Notification User' 
  AND a.action_name IN ('Display', 'Update')
ON CONFLICT (menu_id, action_id) DO NOTHING;

INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu m
CROSS JOIN action a
WHERE m.menu_name = 'Notification Device' 
  AND a.action_name IN ('Display', 'Create','Update','Delete')
ON CONFLICT (menu_id, action_id) DO NOTHING;

-- Create Roles
INSERT INTO role (role_name) 
VALUES ('Admin') 
ON CONFLICT (role_name) DO NOTHING;

-- 2. Grant ALL available menu_actions to the Super Admin
INSERT INTO role_permission (role_id, menu_id, action_id)
SELECT r.role_id, ma.menu_id, ma.action_id
FROM role r
CROSS JOIN menu_action ma
WHERE r.role_name = 'Admin'
ON CONFLICT (role_id, menu_id, action_id) DO NOTHING;