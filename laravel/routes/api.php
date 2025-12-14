<?php

use App\Http\Controllers\AuthController;
use App\Http\Controllers\CommentController;
use App\Http\Controllers\CommonReportController;
use App\Http\Controllers\OnelineController;
use App\Http\Controllers\PageReactionController;
use App\Http\Controllers\PostController;
use App\Http\Controllers\PreviewController;
use App\Http\Controllers\ReplyController;
use App\Http\Controllers\RunboxController;
use App\Http\Controllers\SocialController;
use App\Http\Controllers\UserController;
use App\Http\Controllers\WriteRequestController;
use Illuminate\Support\Facades\Route;

Route::middleware(['mwauth'])->group(function () {
    Route::get('/me', [AuthController::class, 'me']);
    Route::post('/me/avatar', [AuthController::class, 'updateAvatar']);
    Route::get('/me/gravatar/verify', [AuthController::class, 'verifyGravatar']);
    Route::get('/logout', [AuthController::class, 'logout']);
});

Route::get('/user/{userName}', [UserController::class, 'show']);
Route::get('/user/{userId}/stats', [UserController::class, 'stats']);

Route::get('/runbox/{hash}', [RunboxController::class, 'get']);
Route::post('/runbox', [RunboxController::class, 'post']);

Route::get('/comments/recent', [CommentController::class, 'recent']);
Route::get('/comments/{page}', [CommentController::class, 'list']);
Route::post('comments', [CommentController::class, 'store']);
Route::put('/comments/{comment}', [CommentController::class, 'update']);
Route::delete('/comments/{comment}', [CommentController::class, 'destroy']);

Route::get('/reactions/page/{page}', [PageReactionController::class, 'show']);
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

Route::prefix('runbox')->group(function () {
    Route::post('/', [RunboxController::class, 'store']);
    Route::get('{hash}', [RunboxController::class, 'show']);
    Route::post('{hash}/rerun', [RunboxController::class, 'rerun']);
});

Route::prefix('common-report')->group(function () {
    Route::get('/', [CommonReportController::class, 'index']);
    Route::post('/', [CommonReportController::class, 'store']);
    Route::get('{id}', [CommonReportController::class, 'show']);
    Route::delete('{id}', [CommonReportController::class, 'destroy']);
    Route::post('{id}/rerun', [CommonReportController::class, 'rerun']);
    Route::post('{id}/clone', [CommonReportController::class, 'clone']);
});

Route::post('/write-request', [WriteRequestController::class, 'store']);
Route::get('/write-request/todo', [WriteRequestController::class, 'indexTodo']);
Route::get('/write-request/todo-top', [WriteRequestController::class, 'indexTodoTop']);
Route::get('/write-request/done', [WriteRequestController::class, 'indexDone']);
Route::get('/write-request/count', [WriteRequestController::class, 'count']);
Route::delete('/write-request/{id}', [WriteRequestController::class, 'destroy']);
