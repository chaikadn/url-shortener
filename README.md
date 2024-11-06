# url-shortener

# REST API

GET /{id} -- get long url from id       -- 307, 4xx, 500, Location: 'long url'
POST /    -- shorten long url from Body -- 201, 4xx, Body: 'short url'
