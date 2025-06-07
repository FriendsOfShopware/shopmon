UPDATE user
SET notifications = (
    SELECT json_group_array('shop-' || shop.id)
    FROM shop
    INNER JOIN member ON member.organization_id = shop.organization_id
    WHERE member.user_id = user.id
)
WHERE EXISTS (
    SELECT 1
    FROM member
    WHERE member.user_id = user.id
);
