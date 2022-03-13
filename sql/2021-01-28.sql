-- To be able to add foreign key constraints

DELETE FROM categories WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM categories t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.id IS NULL) tt);

DELETE FROM images WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM images t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.id IS NULL) tt);

DELETE FROM prices WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM prices t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.id IS NULL) tt);

DELETE FROM product_price_changes WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM product_price_changes t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.id IS NULL) tt);

DELETE FROM product_view_counts WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM product_view_counts t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.id IS NULL) tt);

DELETE FROM specs WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM specs t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.id IS NULL) tt);

DELETE FROM stocks WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM stocks t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.id IS NULL) tt);

DELETE FROM watch_products WHERE id IN (SELECT tt.id FROM (SELECT t1.id FROM watch_products t1 LEFT JOIN products t2 ON t2.id = t1.product_id WHERE t2.id IS NULL) tt);
