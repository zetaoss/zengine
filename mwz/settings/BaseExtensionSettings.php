<?php

// bundled
wfLoadExtension('CategoryTree');
wfLoadExtension('Cite');
wfLoadExtension('CiteThisPage');
wfLoadExtension('CodeEditor');
wfLoadExtension('ConfirmEdit');
wfLoadExtension('DiscussionTools');
wfLoadExtension('Echo');
wfLoadExtension('Gadgets');
wfLoadExtension('ImageMap');
wfLoadExtension('InputBox');
wfLoadExtension('Interwiki');
wfLoadExtension('Linter');
wfLoadExtension('LoginNotify');
wfLoadExtension('MultimediaViewer');
wfLoadExtension('Nuke');
wfLoadExtension('OATHAuth');
wfLoadExtension('PageImages');
wfLoadExtension('ParserFunctions');
wfLoadExtension('PdfHandler');
wfLoadExtension('Poem');
wfLoadExtension('ReplaceText');
wfLoadExtension('Scribunto');
wfLoadExtension('SecureLinkFixer');
wfLoadExtension('SpamBlacklist');
wfLoadExtension('SyntaxHighlight_GeSHi');
wfLoadExtension('TemplateData');
wfLoadExtension('TextExtracts');
wfLoadExtension('Thanks');
wfLoadExtension('TitleBlacklist');
wfLoadExtension('VisualEditor');
wfLoadExtension('WikiEditor');

// extra
wfLoadExtension('CharInsert');
wfLoadExtension('CrawlerProtection');
wfLoadExtension('EmbedVideo');
wfLoadExtension('MsUpload');
wfLoadExtension('SimpleMaps');
wfLoadExtension('SimpleMarkdown');
wfLoadExtension('SimpleMathJax');
wfLoadExtension('SimpleMermaid');
wfLoadExtension('TemplateStyles');
wfLoadExtension('ZetaExtension');

if (getenv('MW_OAUTH_ENABLED') === '1') {
    wfLoadExtension('OAuth');
    $wgOAuth2PrivateKey = getenv('MW_OAUTH2_PRIVATE_KEY') ?: '';
    $wgOAuth2PublicKey = getenv('MW_OAUTH2_PUBLIC_KEY') ?: '';
    if ($wgOAuth2PrivateKey === '' || $wgOAuth2PublicKey === '') {
        throw new RuntimeException('OAuth requires MW_OAUTH2_PRIVATE_KEY and MW_OAUTH2_PUBLIC_KEY');
    }
    $wgGroupPermissions['user']['mwoauthproposeconsumer'] = true;
    $wgGroupPermissions['user']['mwoauthupdateownconsumer'] = true;
}

$wgVisualEditorUseSingleEditTab            = true;
$wgDefaultUserOptions['visualeditor-tabs'] = 'prefer-wt';
$wgHiddenPrefs[]                           = 'visualeditor-tabs';

$wgScribuntoDefaultEngine = 'luasandbox';

