# Rate Limiter implemented in go
Mostly in reference to https://bytebytego.com/courses/system-design-interview/design-a-rate-limiter

Its a Bad practice to implement the rate limiter on the client side as clientcan forge / modify the request count and is generally considered unreliable.
Better way would be to put it on the server side as part of the existing application code or even better as a seperate middleware, in which case the request is prevented to reach the application.


### Popular Algorithm for Rate Limiting
- Token bucket
- Leaking bucket
- Fixed window counter
- Sliding window log
- Sliding window counter
