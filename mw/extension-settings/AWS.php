<?php

// AWS: S3 as the file backend for uploads, thumbnails and Score renderings.
// Credentials, region and bucket come from the environment (see docs/config.md).
$wgAWSCredentials = [
    'key'    => getenv('AWS_ACCESS_KEY_ID'),
    'secret' => getenv('AWS_SECRET_ACCESS_KEY'),
    'token'  => false,
];
$wgAWSRegion = getenv('AWS_DEFAULT_REGION');

$zetaS3Bucket = getenv('AWS_BUCKET');
$zetaS3Url    = "https://{$zetaS3Bucket}.s3.amazonaws.com";

$wgFileBackends['s3']['containerPaths'] = [
    'zetawiki-local-public'  => $zetaS3Bucket,
    'zetawiki-local-thumb'   => "{$zetaS3Bucket}/thumb",
    'zetawiki-local-deleted' => "{$zetaS3Bucket}/deleted",
    'zetawiki-local-temp'    => "{$zetaS3Bucket}/temp",
    'zetawiki-score-render'  => "{$zetaS3Bucket}/lilypond/score-render",
];
$wgLocalFileRepo = [
    'class'      => 'LocalRepo',
    'name'       => 'local',
    'backend'    => 'AmazonS3',
    'url'        => '/w/img_auth.php',
    'hashLevels' => 2,
    'zones'      => [
        'public'       => ['url' => $zetaS3Url],
        'thumb'        => ['url' => "{$zetaS3Url}/thumb"],
        'temp'         => ['url' => false],
        'deleted'      => ['url' => false],
        'score-render' => ['url' => "{$zetaS3Url}/lilypond/score-render"],
    ],
];
unset($zetaS3Bucket, $zetaS3Url);
