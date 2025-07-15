/* @name get */
SELECT * FROM "zaken" as z
LIMIT :limit! OFFSET :offset!;