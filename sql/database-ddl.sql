CREATE TABLE IF NOT EXISTS twitter_tweet (
    id BIGINT PRIMARY KEY,
    created_at TIMESTAMPTZ,
    screename VARCHAR(100),
    text TEXT,
    tags VARCHAR(100)[]
);

CREATE TABLE IF NOT EXISTS twitter_tweet_indexes (
    id SERIAL PRIMARY KEY,
    from_timestamp TIMESTAMPTZ,
    to_timestamp TIMESTAMPTZ,
    UNIQUE (from_timestamp, to_timestamp)
);

--
-- Trigger to create new indexes when a new row is added to twitter_tweet_indexes
--
CREATE OR REPLACE FUNCTION create_twitter_tweet_index() RETURNS TRIGGER
LANGUAGE 'plpgsql'
AS $_$
DECLARE
    idx_name VARCHAR;
BEGIN
    idx_name := quote_ident('twitter_tweet_index_' || NEW.id::varchar);
    EXECUTE format(
        'CREATE INDEX %I ON twitter_tweet(created_at) '
        'WHERE created_at BETWEEN %L AND %L',
        idx_name, NEW.from_timestamp, NEW.to_timestamp
    );
    RETURN NULL;
END;
$_$;

DROP TRIGGER IF EXISTS create_twitter_tweet_index_trigger ON twitter_tweet_indexes;
CREATE TRIGGER create_twitter_tweet_index_trigger 
AFTER INSERT ON twitter_tweet_indexes
FOR EACH ROW EXECUTE PROCEDURE create_twitter_tweet_index();

--
-- Trigger to delete trigger when removing row from twitter_tweet_indexes
--
CREATE OR REPLACE FUNCTION drop_twitter_tweet_index() RETURNS TRIGGER
LANGUAGE 'plpgsql'
AS $_$
DECLARE
    idx_name VARCHAR;
BEGIN
    idx_name := quote_ident('twitter_tweet_index_' || OLD.id::varchar);
    EXECUTE format(
        'DROP INDEX %I', idx_name
    );
    RETURN NULL;
END;
$_$;

DROP TRIGGER IF EXISTS drop_twitter_tweet_index_trigger ON twitter_tweet_indexes;
CREATE TRIGGER drop_twitter_tweet_index_trigger 
AFTER DELETE ON twitter_tweet_indexes
FOR EACH ROW EXECUTE PROCEDURE drop_twitter_tweet_index();


--
-- Trigger to prevent updates to twitter_tweet_indexes
--
CREATE OR REPLACE FUNCTION update_twitter_tweet_index() RETURNS TRIGGER
LANGUAGE 'plpgsql'
AS $_$
BEGIN
    IF NEW.from_timestamp != OLD.from_timestamp OR NEW.to_timestamp != OLD.to_timestamp THEN
        RAISE EXCEPTION 'Cannot update indexes, you can only create or drop indexes';
    END IF;
END;
$_$;

DROP TRIGGER IF EXISTS update_twitter_tweet_index_trigger ON twitter_tweet_indexes;
CREATE TRIGGER update_twitter_tweet_index_trigger 
BEFORE UPDATE ON twitter_tweet_indexes
FOR EACH ROW EXECUTE PROCEDURE update_twitter_tweet_index();