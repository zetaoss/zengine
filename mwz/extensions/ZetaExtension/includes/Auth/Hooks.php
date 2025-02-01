<?php

namespace ZetaExtension\Auth;

class Hooks
{
    public static function onSpecialPageBeforeExecute($special, $subPage)
    {
        if (get_class($special) != 'SpecialUserLogin') {
            return;
        }
        $returnto = $_GET['returnto'] ?? false;
        if (! $returnto) {
            header('Location: /login');
            exit;
        }
        header("Location: /login?returnto=$returnto");
        exit;
    }

    public static function onPostLoginRedirect($returnTo, $returnToQuery, $type)
    {
        return self::__returnTo();
    }

    public static function onUserLogoutComplete($user, $inject_html, $old_name)
    {
        return self::__returnTo();
    }

    private static function __returnTo()
    {
        $returnto = $_GET['returnto'] ?? false;
        if (! $returnto) {
            return true;
        }
        global $wgZetaAllowHosts;
        if (! in_array(parse_url($returnto, PHP_URL_HOST), $wgZetaAllowHosts)) {
            return true;
        }
        header('Location: '.$returnto);
        exit;
    }
}
