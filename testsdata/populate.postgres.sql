-- sheets
INSERT INTO sheets (name)
VALUES ('single shots');
INSERT INTO sheets (name)
VALUES ('double shots');
INSERT INTO sheets (name)
VALUES ('long blacks');
INSERT INTO sheets (name)
VALUES ('lattes');
-- beans
INSERT INTO beans (
        roaster_name,
        beans_name,
        roast_date,
        roast_level
    )
VALUES (
        'Coffee Collective',
        'Kieni',
        '2023-12-29',
        '0'
    );
INSERT INTO beans (
        roaster_name,
        beans_name,
        roast_date,
        roast_level
    )
VALUES (
        'Nordic Roasting',
        'Vikings do it better',
        '2024-01-01',
        '1'
    );
-- shots
INSERT INTO shots (
        grind_setting,
        quantity_in,
        quantity_out,
        shot_time,
        sheet_id,
        beans_id
    )
VALUES (
        12,
        18.0,
        36.0,
        25,
        2,
        1
    );
-- results
INSERT INTO results (
    rating,
    is_too_bitter,
    is_too_sour,
    comparaison_with_previous_result,
    additional_notes,
    shot_id
  )
VALUES (
    6.3,
    false,
    true,
    0,
    'Lets try more output',
    1
  );