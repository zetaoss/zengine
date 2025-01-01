<?php

use App\Http\Controllers\AuthController;
use App\Http\Controllers\CommentController;
use App\Http\Controllers\CommonReportController;
use App\Http\Controllers\OnelineController;
use App\Http\Controllers\PageReactionController;
use App\Http\Controllers\PostController;
use App\Http\Controllers\PreviewController;
use App\Http\Controllers\ReplyController;
use App\Http\Controllers\RunBoxController;
use App\Http\Controllers\SocialController;
use App\Http\Controllers\WriteRequestController;
use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    return view('welcome');
});

Route::get('/auth', function () {
    Route::get('/redirect/{provider}', [SocialController::class, 'redirect'])->where('provider', 'github|google');
    Route::get('/callback/{provider}', [SocialController::class, 'callback'])->where('provider', 'github|google');
    Route::get('/config-services', function () {
        $services = config('services');
        dd($services);
    });
});

Route::get('/api', function () {
    Route::get('/me', [AuthController::class, 'me']);
    Route::get('/logout', [AuthController::class, 'logout']);

    Route::get('/box/run/{pageid}/{hash}', [RunBoxController::class, 'get']);
    Route::post('/box/run/{pageid}/{hash}', [RunBoxController::class, 'post']);
    Route::post('/box/run/{pageid}/{hash}', [RunBoxController::class, 'put']);

    Route::get('/comments/recent', [CommentController::class, 'recent']);
    Route::get('/comments/{pageID}', [CommentController::class, 'list']);
    Route::post('comments', [CommentController::class, 'store']);
    Route::put('/comments/{comment}', [CommentController::class, 'update']);
    Route::delete('/comments/{comment}', [CommentController::class, 'destroy']);

    Route::get('/catcomments/{pageID}', [CommentController::class, 'catcomments']);
    Route::get('/catcomments-cat/{cat}', [CommentController::class, 'catcommentsCat'])->where('cat', '(.*)');

    Route::get('/reactions/page/{pageID}', [PageReactionController::class, 'show']);
    Route::post('/reactions/page', [PageReactionController::class, 'store']);

    Route::get('/onelines/recent', [OnelineController::class, 'recent']);

    Route::get('/posts/recent', [PostController::class, 'recent']);

    Route::get('/posts', [PostController::class, 'index']);
    Route::post('/posts', [PostController::class, 'store']);
    Route::get('/posts/{post}', [PostController::class, 'show']);
    Route::put('/posts/{post}', [PostController::class, 'update']);
    Route::delete('/posts/{post}', [PostController::class, 'destroy']);

    Route::get('/posts/{post}/replies', [ReplyController::class, 'index']);
    Route::post('/posts/{post}/replies', [ReplyController::class, 'store']);
    Route::put('/posts/{post}/replies/{reply}', [ReplyController::class, 'update']);
    Route::delete('/posts/{post}/replies/{reply}', [ReplyController::class, 'destroy']);

    Route::get('/preview', [PreviewController::class, 'show']);
    Route::get('/auth/social/check/{code}', [SocialController::class, 'checkCode']);
    Route::get('/auth/social/login/{code}', [SocialController::class, 'loginCode']);

    Route::get('/common-report', [CommonReportController::class, 'index']);
    Route::get('/common-report/{id}', [CommonReportController::class, 'show']);
    Route::post('/common-report', [CommonReportController::class, 'store']);
    Route::delete('/common-report/{id}', [CommonReportController::class, 'destroy']);

    Route::post('/write-request', [WriteRequestController::class, 'store']);
    Route::get('/write-request/todo', [WriteRequestController::class, 'indexTodo']);
    Route::get('/write-request/todo-top', [WriteRequestController::class, 'indexTodoTop']);
    Route::get('/write-request/done', [WriteRequestController::class, 'indexDone']);
    Route::get('/write-request/count', [WriteRequestController::class, 'count']);
    Route::delete('/write-request/{id}', [WriteRequestController::class, 'destroy']);
});
