table "shop" {
  schema = schema.shopmon
  column "id" {
    null           = false
    type           = int
    auto_increment = true
  }
  column "team_id" {
    null = false
    type = int
  }
  column "name" {
    null = false
    type = varchar(255)
  }
  column "status" {
    null    = false
    type    = varchar(255)
    default = "green"
  }
  column "url" {
    null = false
    type = varchar(255)
  }
  column "favicon" {
    null = true
    type = varchar(500)
  }
  column "client_id" {
    null = false
    type = varchar(255)
  }
  column "client_secret" {
    null = false
    type = varchar(255)
  }
  column "shopware_version" {
    null = true
    type = varchar(255)
  }
  column "last_scraped_at" {
    null = true
    type = datetime
  }
  column "last_scraped_error" {
    null = true
    type = text
  }
  column "ignores" {
    null    = true
    type    = text
    default = sql("(_utf8mb4'[]')")
  }
  column "shop_image" {
    null = true
    type = varchar(128)
  }
  column "created_at" {
    null    = false
    type    = datetime
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null = true
    type = datetime
  }
  primary_key {
    columns = [column.id]
  }
  index "url" {
    unique  = true
    columns = [column.url]
  }
}
table "shop_changelog" {
  schema = schema.shopmon
  column "id" {
    null           = false
    type           = int
    auto_increment = true
  }
  column "shop_id" {
    null = false
    type = int
  }
  column "extensions" {
    null = true
    type = text
  }
  column "old_shopware_version" {
    null = true
    type = varchar(255)
  }
  column "new_shopware_version" {
    null = true
    type = varchar(255)
  }
  column "date" {
    null = true
    type = datetime
  }
  primary_key {
    columns = [column.id]
  }
}
table "shop_pagespeed" {
  schema = schema.shopmon
  column "id" {
    null           = false
    type           = int
    auto_increment = true
  }
  column "shop_id" {
    null    = false
    type    = int
    default = 0
  }
  column "created_at" {
    null    = false
    type    = datetime
    default = sql("CURRENT_TIMESTAMP")
  }
  column "performance" {
    null    = false
    type    = tinyint
    default = 0
  }
  column "accessibility" {
    null    = false
    type    = tinyint
    default = 0
  }
  column "bestpractices" {
    null    = false
    type    = tinyint
    default = 0
  }
  column "seo" {
    null    = false
    type    = tinyint
    default = 0
  }
  primary_key {
    columns = [column.id]
  }
}
table "shop_scrape_info" {
  schema = schema.shopmon
  column "shop_id" {
    null = false
    type = int
  }
  column "extensions" {
    null = false
    type = text
  }
  column "scheduled_task" {
    null = false
    type = text
  }
  column "queue_info" {
    null = false
    type = text
  }
  column "cache_info" {
    null = false
    type = text
  }
  column "checks" {
    null = true
    type = text
  }
  column "created_at" {
    null = false
    type = datetime
  }
  primary_key {
    columns = [column.shop_id]
  }
}
table "team" {
  schema = schema.shopmon
  column "id" {
    null           = false
    type           = int
    auto_increment = true
  }
  column "name" {
    null = true
    type = varchar(255)
  }
  column "owner_id" {
    null = false
    type = int
  }
  column "created_at" {
    null    = false
    type    = datetime
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null = true
    type = datetime
  }
  primary_key {
    columns = [column.id]
  }
  index "name" {
    unique  = true
    columns = [column.name]
  }
}
table "user" {
  schema = schema.shopmon
  column "id" {
    null           = false
    type           = int
    auto_increment = true
  }
  column "username" {
    null = true
    type = varchar(50)
  }
  column "email" {
    null = false
    type = varchar(255)
  }
  column "password" {
    null = false
    type = varchar(255)
  }
  column "verify_code" {
    null = true
    type = varchar(32)
  }
  column "created_at" {
    null    = true
    type    = datetime
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null = true
    type = datetime
  }
  primary_key {
    columns = [column.id]
  }
  index "email" {
    unique  = true
    columns = [column.email]
  }
  index "verify_code" {
    unique  = true
    columns = [column.verify_code]
  }
}
table "user_notification" {
  schema = schema.shopmon
  column "id" {
    null           = false
    type           = int
    auto_increment = true
  }
  column "user_id" {
    null = false
    type = int
  }
  column "key" {
    null = false
    type = varchar(50)
  }
  column "level" {
    null = false
    type = varchar(10)
  }
  column "title" {
    null    = false
    type    = varchar(128)
    default = ""
  }
  column "message" {
    null = false
    type = varchar(128)
  }
  column "link" {
    null = false
    type = varchar(128)
  }
  column "read" {
    null    = false
    type    = tinyint
    default = 0
  }
  column "created_at" {
    null    = true
    type    = datetime
    default = sql("CURRENT_TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "unique_notification" {
    unique  = true
    columns = [column.user_id, column.key]
  }
}
table "user_to_team" {
  schema = schema.shopmon
  column "user_id" {
    null = false
    type = int
  }
  column "team_id" {
    null = false
    type = int
  }
  index "user_id_team_id" {
    unique  = true
    columns = [column.user_id, column.team_id]
  }
}
schema "shopmon" {
  charset = "utf8mb4"
  collate = "utf8mb4_0900_ai_ci"
}
