alter table reply rename column reply_id to id;
alter table reply rename column reply_idempotency_key to idempotency_key;
alter table reply rename column reply_signature to signature;
alter table reply rename column reply_article to article;
alter table reply rename column reply_body to body;
alter table reply rename column reply_deleted to deleted;
alter table reply rename column reply_created_at to created_at;
alter table reply rename column reply_author_name to author_name;
