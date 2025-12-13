<?php

namespace ZetaExtension\Auth;

class Hooks
{
    private static function redirectToSpa($special, string $path): bool
    {
        $req = $special->getRequest();
        $returnto = $req->getVal('returnto', '');

        $target = $path;
        if ($returnto !== '') {
            $target .= '?returnto='.rawurlencode($returnto);
        }

        $special->getOutput()->redirect($target);

        return false;
    }

    public static function onSpecialPageBeforeExecute($special, $subPage)
    {
        if ($special instanceof \SpecialUserLogin) {
            return self::redirectToSpa($special, '/login');
        }

        if ($special instanceof \SpecialUserLogout) {
            return self::redirectToSpa($special, '/logout');
        }

        return true;
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
