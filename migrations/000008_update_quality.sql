-- +goose Up
UPDATE places SET quality=40 WHERE photo_url='5k-alfabank-2.jpg';
UPDATE places SET quality=35 WHERE photo_url='5k-alfabank-3.jpg';
UPDATE places SET quality=35 WHERE photo_url='3a-cafe-2.jpg';

-- +goose Down
UPDATE places SET quality=100 WHERE photo_url='5k-alfabank-2.jpg';
UPDATE places SET quality=100 WHERE photo_url='5k-alfabank-3.jpg';
UPDATE places SET quality=50 WHERE photo_url='3a-cafe-2.jpg';
