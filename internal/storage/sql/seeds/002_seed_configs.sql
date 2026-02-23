INSERT INTO configs (website_id, key, value, status, created_at, updated_at)
SELECT
    w.id,
    'scraper_config',
    '{
        "domain": "badomjip.com",
        "start_url": "https://badomjip.com/",
        "product_list": {
            "item": "div.product-grid-item",
            "fields": {
                "title": {
                    "selector": "h3.wd-entities-title a",
                    "attr": ""
                },
                "price": {
                    "selector": "span.price",
                    "attr": ""
                },
                "link": {
                    "selector": "a.product-image-link",
                    "attr": "href"
                },
                "image": {
                    "selector": "a.product-image-link img",
                    "attr": "src"
                }
            }
        },
        "pagination": {
            "next": "a.next.page-numbers"
        },
        "product_detail": {
            "description": {
                "selector": "div.woocommerce-product-details__short-description, div.woocommerce-Tabs-panel--description"
            }
        }
    }'::json,
    'active',
    NOW(),
    NOW()
FROM websites w
WHERE w.domain = 'badomjip.com'
ON CONFLICT DO NOTHING;

INSERT INTO configs (website_id, key, value, status, created_at, updated_at)
SELECT
    w.id,
    'scraper_config',
    '{
        "domain": "hosseinibrothers.ir",
        "start_url": "https://hosseinibrothers.ir/",
        "product_list": {
            "item": "div.js-product-miniature",
            "fields": {
                "title": {
                    "selector": "a.stsb_mini_product_name",
                    "attr": ""
                },
                "price": {
                    "selector": "div.stsb_pm_price",
                    "attr": ""
                },
                "link": {
                    "selector": "a.stsb_mini_product_name",
                    "attr": "href"
                },
                "image": {
                    "selector": "img.stsb_pm_image",
                    "attr": "src"
                }
            }
        },
        "pagination": {
            "next": "a[rel=''next'']"
        },
        "product_detail": {
            "description": {
                "selector": "div.stsb_pro_summary p, div.stsb_read_more_box p"
            }
        }
    }'::json,
    'active',
    NOW(),
    NOW()
FROM websites w
WHERE w.domain = 'hosseinibrothers.ir'
ON CONFLICT DO NOTHING;