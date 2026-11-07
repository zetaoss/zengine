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
wfLoadExtension('OAuth');
wfLoadExtension('SimpleMaps');
wfLoadExtension('SimpleMarkdown');
wfLoadExtension('SimpleMathJax');
wfLoadExtension('SimpleMermaid');
wfLoadExtension('TemplateStyles');
wfLoadExtension('ZetaExtension');

$wgOAuth2PrivateKey = 'file:///run/secrets/mediawiki-oauth/private.pem';
$wgOAuth2PublicKey = 'file:///run/secrets/mediawiki-oauth/public.pem';
$wgGroupPermissions['user']['mwoauthproposeconsumer'] = true;
$wgGroupPermissions['user']['mwoauthupdateownconsumer'] = true;

$wgVisualEditorUseSingleEditTab            = true;
$wgDefaultUserOptions['visualeditor-tabs'] = 'prefer-wt';
$wgHiddenPrefs[]                           = 'visualeditor-tabs';

$wgScribuntoDefaultEngine = 'luasandbox';

