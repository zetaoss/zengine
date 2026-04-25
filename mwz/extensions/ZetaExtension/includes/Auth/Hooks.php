<?php

namespace ZetaExtension\Auth;

class Hooks
{
    private static function redirectWithReturnTo($special, string $path): bool
    {
        $req = $special->getRequest();
        $returnto = $req->getVal('returnto', '');
        $returntoquery = $req->getVal('returntoquery', '');

        $query = [];
        if ($returnto !== '') {
            $query['returnto'] = $returnto;
        }
        if ($returntoquery !== '') {
            $query['returntoquery'] = $returntoquery;
        }

        $target = $path;
        if (! empty($query)) {
            $separator = str_contains($path, '?') ? '&' : '?';
            $target .= $separator.http_build_query($query);
        }

        $special->getOutput()->redirect($target);

        return false;
    }

    public static function onSpecialPageBeforeExecute($special, $subPage)
    {
        if ($special instanceof \SpecialUserLogin) {
            return self::redirectWithReturnTo($special, '/login');
        }

        if ($special instanceof \SpecialUserLogout) {
            return self::redirectWithReturnTo($special, '/logout');
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
        $target = (string) ($_GET['returnto'] ?? '');
        if (str_starts_with($target, ':/') && ! str_starts_with($target, '://')) {
            header('Location: '.substr($target, 1));
            exit;
        }

        return true;
    }
}
