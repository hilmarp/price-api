-- Remove ?add-to-cart from nexus.is

UPDATE products SET url = REGEXP_REPLACE(url, '\\?add-to-cart=[0-9]+', '') WHERE source = 'nexus.is' AND url LIKE '%?add-to-cart=%';

DELETE FROM products WHERE source = 'nexus.is' AND url LIKE '%?add-to-cart=%';
