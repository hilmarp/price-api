-- Remove ?vid=123 from heimkaup.is URLs

UPDATE IGNORE products SET url = REGEXP_REPLACE(url, '\\?vid=[0-9]+', '') WHERE source = 'heimkaup.is' AND url LIKE '%?vid=%';

DELETE FROM categories WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM categories t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.source = 'heimkaup.is' AND t2.url LIKE '%?vid=%') tt);

DELETE FROM images WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM images t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.source = 'heimkaup.is' AND t2.url LIKE '%?vid=%') tt);

DELETE FROM prices WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM prices t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.source = 'heimkaup.is' AND t2.url LIKE '%?vid=%') tt);

DELETE FROM product_price_changes WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM product_price_changes t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.source = 'heimkaup.is' AND t2.url LIKE '%?vid=%') tt);

DELETE FROM product_view_counts WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM product_view_counts t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.source = 'heimkaup.is' AND t2.url LIKE '%?vid=%') tt);

DELETE FROM specs WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM specs t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.source = 'heimkaup.is' AND t2.url LIKE '%?vid=%') tt);

DELETE FROM stocks WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM stocks t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.source = 'heimkaup.is' AND t2.url LIKE '%?vid=%') tt);

DELETE FROM watch_products WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM watch_products t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.source = 'heimkaup.is' AND t2.url LIKE '%?vid=%') tt);

DELETE FROM products WHERE source = 'heimkaup.is' AND url LIKE '%?vid=%';
