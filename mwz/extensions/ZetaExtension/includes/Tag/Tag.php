<?php
namespace ZetaExtension\Tag;

class Tag
{
    public static function onParserFirstCallInit($parser)
    {
        $parser->setHook('tvpot', [self::class, 'renderTvpot']);
        $parser->setHook('facebook', [self::class, 'renderFacebook']);
        $parser->setHook('bgmstore', [self::class, 'renderBgmstore']);
        $parser->setHook('vimeo', [self::class, 'renderVimeo']);
        $parser->setHook('ted', [self::class, 'renderTed']);

        $parser->setHook('jsbin', [self::class, 'renderJsbin']);
        $parser->setHook('codepen', [self::class, 'renderCodepen']);
        $parser->setHook('jsfiddle', [self::class, 'renderJsfiddle']);
        $parser->setHook('codesandbox', [self::class, 'renderCodesandbox']);
        $parser->setHook('gspreadsheets', [self::class, 'renderGspreadsheets']);
        $parser->setFunctionHook('img', [self::class, 'renderImg']);
    }

    public static function renderImg($parser, $param1 = '', $param2 = '', $param3 = '')
    {
        if (filter_var($param1, FILTER_VALIDATE_URL) == false) {
            return ["<span class='border'>NOT VALID URL</span>", 'noparse' => true, 'isHTML' => true];
        }
        $width = is_numeric($param2) ? $param2 . 'px' : 'auto';
        $class = preg_match('/^[a-z- ]+$/iD',$param3) ? $param3 : '';
        return ["<img src='$param1' style='width:$width' class='$class' />", 'noparse' => true, 'isHTML' => true];
    }

    public static function renderTvpot($input, $args, $parser)
    {
        $vid = '';
        $width = $width_max = 640; //425;
        $height = $height_max = 360; //355;
        if (!empty($input)) {
            $vid = $input;
        }
        if (!empty($args['width']) && settype($args['width'], 'integer') && ($width_max >= $args['width'])) {
            $width = $args['width'];
        }
        if (!empty($args['height']) && settype($args['height'], 'integer') && ($height_max >= $args['height'])) {
            $height = $args['height'];
        }
        if (empty($vid)) {
            return;
        }
        return "<div style='width:${width}px;max-width:100%;'>
<div style='position:relative;padding-bottom:56.25%;'>
<iframe style='position:absolute;width:100%;height:100%;'
src='http://videofarm.daum.net/controller/video/viewer/Video.html?vid=$vid&play_loc=daum_tistory'
frameborder='0'
scrolling='no'></iframe>
</div></div>";
    }

    public static function renderFacebook($input, $args, $parser)
    {
        $url = '';
        $width = $width_max = 640; //425;
        $height = $height_max = 360; //355;

        if (!empty($input)) {
            $url = $input;
        }
        if (!empty($args['width']) && settype($args['width'], 'integer') && ($width_max >= $args['width'])) {
            $width = $args['width'];
        }
        if (!empty($args['height']) && settype($args['height'], 'integer') && ($height_max >= $args['height'])) {
            $height = $args['height'];
        }
        if (!empty($url) && filter_var($url, FILTER_VALIDATE_URL) !== false) {
            return "<iframe class='fb' data-url='$url' allowfullscreen='true'></iframe>";
        }
    }

    public static function renderBgmstore($input, $args, $parser)
    {
        $vid = '';
        $width = $width_max = 422; //425;
        $height = $height_max = 180; //355;

        if (!empty($input)) {
            $vid = $input;
        }
        if (!empty($args['width']) && settype($args['width'], 'integer') && ($width_max >= $args['width'])) {
            $width = $args['width'];
        }
        if (!empty($args['height']) && settype($args['height'], 'integer') && ($height_max >= $args['height'])) {
            $height = $args['height'];
        }
        if (!empty($vid)) {
            return "<div style='width:${width}px;max-width:100%;'>
<embed src='http://player.bgmstore.net/$vid' allowscriptaccess='always' allowfullscreen='true' width='422' height='180'></embed>
</div></div>";
        }
    }

    public static function renderVimeo($input, $args, $parser)
    {
        $vid = '';
        $width = $width_max = 640; //425;
        $height = $height_max = 360; //355;
        if (!empty($input)) {
            $vid = $input;
        }
        if (!empty($args['width']) && settype($args['width'], 'integer') && ($width_max >= $args['width'])) {
            $width = $args['width'];
        }
        if (!empty($args['height']) && settype($args['height'], 'integer') && ($height_max >= $args['height'])) {
            $height = $args['height'];
        }
        if (!empty($vid)) {
            return "<div style='width:${width}px;max-width:100%;'>
<div style='position:relative;padding-bottom:56.25%;'>
<iframe style='position:absolute;width:100%;height:100%;border:0' src='https://player.vimeo.com/video/$vid' webkitallowfullscreen mozallowfullscreen allowfullscreen></iframe>
</div></div>";
        }
    }

    public static function renderTed($input, $args, $parser)
    {
        $vid = '';
        $width = $width_max = 640; //425;
        $height = $height_max = 360; //355;
        if (!empty($input)) {
            $vid = $input;
        }
        if (!empty($args['width']) && settype($args['width'], 'integer') && ($width_max >= $args['width'])) {
            $width = $args['width'];
        }
        if (!empty($args['height']) && settype($args['height'], 'integer') && ($height_max >= $args['height'])) {
            $height = $args['height'];
        }
        if (!empty($vid)) {
            return "<div style='width:${width}px;max-width:100%;'>
<div style='position:relative;padding-bottom:56.25%'>
<iframe style='position:absolute;width:100%;height:100%;'
src='//embed.ted.com/talks/lang/ko/$vid.html'
frameborder='0'
scrolling='no'></iframe>
</div></div>";
        }
    }

    public static function renderJsbin($input, $args, $parser)
    {
        $id = '';
        $embed = 'console,output';
        $height = 120;
        if (!empty($input)) {
            $id = $input;
        }
        if (!empty($args['height']) && settype($args['height'], 'integer')) {
            $height = $args['height'] + 36;
        }
        if (!empty($args['embed'])) {
            if (settype($args['embed'], 'string')) {
                $embed = $args['embed'];
            }
        }
        if (!empty($id)) {
            return "<div style='border:0;border-left:6px solid silver;width:calc( 100% - 6px );'>
<a class='jsbin-embed' href='http://jsbin.com/$id/embed?$embed'> on jsbin.com</a>
</div>
<script src='http://static.jsbin.com/js/embed.min.js?3.34.1'></script>
";
        }
    }

    public static function renderCodepen($input, $args, $parser)
    {
        $id = '';
        $styles = '';
        $height = $height_min = 160;
        if (!empty($input)) {
            $id = $input;
        }
        if (!empty($args['height']) && settype($args['height'], 'integer') && ($height_min <= $args['height'])) {
            $height = $args['height'];
        }
        if (!empty($id)) {
            return "<div style='border:1px solid #3D3D3E;border-top:0'>
<p data-height='${height}' data-theme-id='0' data-slug-hash='$id' data-default-tab='result' class='codepen'>
</div>
<script async src='//assets.codepen.io/assets/embed/ei.js'></script>";
        }
    }

    public static function renderJsfiddle($input, $args, $parser)
    {
        $vid = '';
        $styles = 'result';
        $height = 120;
        if (!empty($input)) {
            $vid = $input;
        }
        if (!empty($args['height']) && settype($args['height'], 'integer')) {
            $height = $args['height'] + 36;
        }
        if (!empty($args['styles']) && settype($args['styles'], 'string')) {
            if ($args['styles'] == 'all') {
                $styles = '';
            } else {
                $styles = $args['styles'];
            }
        }
        if (empty($vid)) {
            return "";
        }
        return str_replace(PHP_EOL, ' ', "<div><iframe class='z-embed' style='height:${height}px'
sandbox='allow-forms allow-popups allow-scripts allow-same-origin allow-modals'
src='//jsfiddle.net/${vid}/embedded/result' frameborder='0'
allowfullscreen='allowfullscreen' scrolling='no'></iframe></div>");
    }

    private static function __fillArgv(&$args, $rules)
    {
        foreach ($rules as $name => $default) {
            if (empty($args[$name]) || !settype($args[$name], gettype($default))) {
                $args[$name] = $default;
            }

        }
    }

    public static function renderCodesandbox($id, $args, $parser)
    {
        if (empty($id)) {
            return '';
        }

        self::__fillArgv($args, [
            'height' => 400,
            'preview-height' => 400,
            'fontsize' => 14,
        ]);
        return "<div><iframe src='https://codesandbox.io/embed/${id}?fontsize=${argv['fontsize']}&hidenavigation=1&view=editor' style='width:100%; height:${argv['height']}px; border:0;' sandbox='allow-forms allow-modals allow-popups allow-presentation allow-same-origin allow-scripts'></iframe><iframe src='https://codesandbox.io/embed/${id}?hidenavigation=1&view=preview' style='width:100%; height:${argv['preview-height']}px; border:0;' allow='accelerometer; ambient-light-sensor; camera; encrypted-media; geolocation; gyroscope; hid; microphone; midi; payment; usb; vr; xr-spatial-tracking' sandbox='allow-forms allow-modals allow-popups allow-presentation allow-same-origin allow-scripts'></iframe></div>";
    }

    public static function renderGspreadsheets($input, $args, $parser)
    {
        $id = '';
        $styles = '';
        $height = $height_min = 200;

        if (!empty($input)) {
            $id = $input;
        }
        if (!empty($args['height']) && settype($args['height'], 'integer') && ($height_min <= $args['height'])) {
            $height = $args['height'];
        }

        if (!empty($id)) {
            return "<div>
<iframe style='border-left:6px solid silver;width:calc( 100% - 6px );height:${height}px'
src='//docs.google.com/spreadsheets/d/$id/pubhtml?gid=0&amp;single=true&amp;widget=false&amp;headers=true'
scrolling='no'></iframe>
</div>";
        }
    }
}
