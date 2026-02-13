-- +goose Up
INSERT INTO meetings (dill_id, doe_id, pair_score, is_fullmatch, place_id, time)
  VALUES (780074874, 1061574811, 1.00, false, 5, '2026-02-14 16:00:00+00');

-- +goose Down
DELETE FROM meetings WHERE dill_id = 780074874 AND doe_id = 1061574811;
