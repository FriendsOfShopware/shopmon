<?php

use Doctrine\DBAL\Connection;
use Shopware\Core\Framework\Context;

$kernel = require '/opt/shopware/boot.php';

/** @var Connection $connection */
$connection = $kernel->getContainer()->get(Connection::class);

$connection->executeStatement("DELETE FROM user");

$scheduledTask = $connection->fetchAllAssociative('SELECT * FROM integration');

if (!empty($scheduledTask)) {
    return;
}

$integration = $kernel->getContainer()->get('integration.repository');

$integration->create([
    [
        'label' => 'Shopmon',
        'accessKey' => 'SWIAUZL4OXRKEG1RR3PMCEVNMG',
        'secretAccessKey' => 'aXhNQ3NoRHZONmxPYktHT0c2c09rNkR0UHI0elZHOFIycjBzWks',
        'admin' => false,
        'aclRoles' => [
            [
                'name' => 'Shopmon',
                'privileges' => [
                    'app:read',
                    'plugin:read',
                    'system_config:read',
                    'scheduled_task:read',
                    'frosh_tools:read',
                    'system:clear:cache',
                    'system:cache:info'
                ]
            ]
        ]
    ]

], Context::createDefaultContext());
