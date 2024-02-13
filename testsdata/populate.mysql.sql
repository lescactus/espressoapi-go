-- sheets
INSERT INTO sheets (name)
VALUES
('single shots'),
('double shots'),
('long blacks'),
('lattes');
-- roasters
INSERT INTO roasters (name)
VALUES
('Coffee Collective'),
('Nordic Roasting');
-- beans
INSERT INTO beans (name, roast_date, roast_level, roaster_id)
VALUES
('Kieni', '2023-12-29', '0', (SELECT id FROM roasters WHERE name = 'Coffee Collective')),
('Vikings do it better', '2024-01-01', '1', (SELECT id FROM roasters WHERE name = 'Nordic Roasting'));
-- shots
-- 0 = false
-- 1 = true
INSERT INTO shots (grind_setting, quantity_in, quantity_out, shot_time, rating, is_too_bitter, is_too_sour, comparaison_with_previous_result, additional_notes, sheet_id, beans_id)
VALUES
(12, 18.0, 36.0, 25, 6.3, 0, 1, 0, 'Lets try more output', 2, 1),
(12, 18.0, 38.0, 26, 8.0, 0, 0, 1, 'Pretty good', 2, 1),
(11, 18.0, 37.5, 25, 7.1, 0, 1, 0, 'Should increase water temperature?', 2, 2);