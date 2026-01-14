⚪︎ 認証は Session Cookie
POST /api/auth/login

⚫︎ 内容:メールとパスワードでログインし、Set-Cookie:session_id を返す。
⚫︎ 認証:不要
⚫︎Body:
・email string
・password string
⚫︎ 成功: 200OK
⚫︎ 失敗:
・400 INVALID_REQUEST 入力不足
・401 UNAUTHETICATED 認証失敗
・500 INTERNAL_ERROR 　サーバーエラー

POST /api/auth/logout
⚫︎ 内容: セッションを無効化し、Cookie を削除する。
⚫︎ 認証: 必要 session cookie
⚫︎ 成功: 200OK
⚫︎ 失敗:
・401 UNAUTHETICATED 未ログインまたはセッション無効
・500 INTERNAL_ERROR サーバーエラー

GET /api/auth/me
⚫︎ 内容: ログイン中のユーザー情報を返す。
⚫︎ 認証: 必要 session cookie
⚫︎ 成功: 200OK
⚫︎ 失敗: 401 UNAUTHETICATED 未ログインまたはセッション無効

---

⚪︎ 商品
POST /api/products

⚫︎ 内容： 商品を作成して、在庫にも個数０で作成する。(トランザクション)
⚫︎ 認証: 必要 session cookie
⚫︎ Body: name string
⚫︎ 成功: 201 Created
⚫︎ 失敗:
・400 INVALID_REQUEST 入力不足
・401 UNAUTHETICATED 認証失敗
・500 INTERNAL_ERROR サーバーエラー

---

⚪︎ 在庫
GET /api/inventories

⚫︎ 内容: 在庫一覧(product と quantity)
⚫︎ 認証: 必要 session cookie
⚫︎ 成功: 200OK
⚫︎ 失敗:
・401 UNAUTHETICATED 認証失敗
・500 INTERNAL_ERROR サーバーエラー

---

⚪︎ 在庫増減
POST /api/stock/inbound

⚫︎ 内容: 在庫の数量を増やす。在庫更新と履歴追加を追加する(トランザクション)。
⚫︎ 認証: 必要 session cookie
⚫︎ Body:
・product_id UUID string
・quantity int 1 以上
・ reason string
⚫︎ 成功: 200OK
⚫︎ 失敗:
・400 INVALID_REQUEST UUID 不正、quantity 不正など。
・401 UNAUTHETICATED 認証失敗
・404 NOT_FOUND 　対象の product や inventory がない。
・500 INTERNAL_ERROR サーバーエラー

POST /api/stock/outbound

⚫︎ 内容: 在庫の数量を減らす。在庫不足は 409。在庫更新と履歴追加を追加する(トランザクション)。
⚫︎ 認証: 必要 session cookie
⚫︎ Body:
・product_id UUID string
・quantity int 1 以上
・ reason string
⚫︎ 成功: 200OK
⚫︎ 失敗:
・400 INVALID_REQUEST UUID 不正、quantity 不正など。
・401 UNAUTHETICATED 認証失敗
・404 NOT_FOUND 　対象の product や inventory がない。
・409 INSUFFICIENT_STOCK 在庫不足
・500 INTERNAL_ERROR サーバーエラー

---

⚪︎ 履歴
GET /api/stock_movements

⚫︎ 内容: 在庫の増減履歴を新しい順番で返す。ページングは limit
⚫︎ 認証: 必要 session cookie
⚫︎ Query:
・ limit 1~200,未指定は 50
・ product_id UUID string 　指定時はその商品のみ。

⚫︎ 成功: 200OK
⚫︎ 失敗:
・ 400 INVALID_REQUEST(product_id が不正、limit が 1〜200 以外)
・ 401 UNAUTHENTICATED 未ログインまたはセッション無効
・ 500 INTERNAL_ERROR サーバーエラー
