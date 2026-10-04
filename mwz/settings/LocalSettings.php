<?php

// MediaWiki entry point. Each file may override values set by the previous one.
//   BaseSettings.php       common settings (this image); values and secrets from environment variables
//   SiteSettings.php       per-deployment settings, provided by the deployment (required)
//   ExtensionSettings.php  all extensions (loading and configuration), provided by the deployment (required)
// See docs/config.md.

if (! defined('MEDIAWIKI')) {
    exit;
}

require "$IP/BaseSettings.php";
require "$IP/SiteSettings.php";
require "$IP/ExtensionSettings.php";
