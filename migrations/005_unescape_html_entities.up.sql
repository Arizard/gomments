UPDATE reply SET body = REPLACE(body, '&#39;', '''');
UPDATE reply SET body = REPLACE(body, '&#34;', '"');
UPDATE reply SET body = REPLACE(body, '&quot;', '"');
UPDATE reply SET body = REPLACE(body, '&amp;', '&');
UPDATE reply SET body = REPLACE(body, '&lt;', '<');
UPDATE reply SET body = REPLACE(body, '&gt;', '>');

UPDATE reply SET author_name = REPLACE(author_name, '&#39;', '''');
UPDATE reply SET author_name = REPLACE(author_name, '&#34;', '"');
UPDATE reply SET author_name = REPLACE(author_name, '&quot;', '"');
UPDATE reply SET author_name = REPLACE(author_name, '&amp;', '&');
UPDATE reply SET author_name = REPLACE(author_name, '&lt;', '<');
UPDATE reply SET author_name = REPLACE(author_name, '&gt;', '>');
