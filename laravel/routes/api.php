<?php

// laravel/routes/api.php

use App\Http\Controllers\AuthController;
use App\Http\Controllers\CommonReportController;
use App\Http\Controllers\OnelineController;
use App\Http\Controllers\PageCommentController;
use App\Http\Controllers\PageReactionController;
use App\Http\Controllers\PostController;
use App\Http\Controllers\PreviewController;
use App\Http\Controllers\ReplyController;
use App\Http\Controllers\RunboxController;
use App\Http\Controllers\UserController;
use App\Http\Controllers\UserProfileController;
use App\Http\Controllers\WriteRequestController;
use Illuminate\Support\Facades\Route;

Route::get('/comments/recent', [PageCommentController::class, 'recent']);
Route::get('/comments/{pageID}', [PageCommentController::class, 'list']);
Route::post('/comments', [PageCommentController::class, 'store'])->middleware('mwauth');
Route::put('/comments/{comment}', [PageCommentController::class, 'update'])->middleware('mwauth');
Route::delete('/comments/{comment}', [PageCommentController::class, 'destroy'])->middleware('mwauth');

Route::get('/common-report', [CommonReportController::class, 'index']);
Route::get('/common-report/{id}', [CommonReportController::class, 'show']);
Route::post('/common-report', [CommonReportController::class, 'store'])->middleware('mwauth');
Route::post('/common-report/{id}/clone', [CommonReportController::class, 'clone'])->middleware('mwauth');
Route::post('/common-report/{id}/rerun', [CommonReportController::class, 'rerun'])->middleware('mwauth');
Route::delete('/common-report/{id}', [CommonReportController::class, 'destroy'])->middleware('mwauth');

Route::get('/me', [AuthController::class, 'me'])->middleware('mwauth:maybe');
Route::get('/me/avatar', [AuthController::class, 'getAvatar'])->middleware('mwauth');
Route::post('/me/avatar', [AuthController::class, 'updateAvatar'])->middleware('mwauth');
Route::get('/me/gravatar/verify', [AuthController::class, 'verifyGravatar'])->middleware('mwauth');

Route::get('/onelines/recent', [OnelineController::class, 'recent']);
Route::get('/onelines', [OnelineController::class, 'index']);
Route::post('/onelines', [OnelineController::class, 'store'])->middleware('mwauth');
Route::delete('/onelines/{oneline}', [OnelineController::class, 'destroy'])->middleware('mwauth');

Route::get('/posts', [PostController::class, 'index']);
Route::get('/posts/recent', [PostController::class, 'recent']);
Route::get('/posts/{post}', [PostController::class, 'show']);
Route::post('/posts', [PostController::class, 'store'])->middleware('mwauth');
Route::put('/posts/{post}', [PostController::class, 'update'])->middleware('mwauth');
Route::delete('/posts/{post}', [PostController::class, 'destroy'])->middleware('mwauth');

Route::get('/posts/{post}/replies', [ReplyController::class, 'index']);
Route::post('/posts/{post}/replies', [ReplyController::class, 'store'])->middleware('mwauth');
Route::put('/posts/{post}/replies/{reply}', [ReplyController::class, 'update'])->middleware('mwauth');
Route::delete('/posts/{post}/replies/{reply}', [ReplyController::class, 'destroy'])->middleware('mwauth');

Route::get('/preview', [PreviewController::class, 'show']);

Route::get('/reactions/page/{page}', [PageReactionController::class, 'show']);
Route::post('/reactions/page', [PageReactionController::class, 'store'])->middleware('mwauth');

Route::get('/runbox/{hash}', [RunboxController::class, 'show']);
Route::post('/runbox', [RunboxController::class, 'store']);
Route::post('/runbox/{hash}/rerun', [RunboxController::class, 'rerun'])->middleware('mwauth');

Route::get('/user/{userId}/stats', [UserController::class, 'stats']);
Route::get('/user/{userName}', [UserController::class, 'show']);

Route::get('/write-request/count', [WriteRequestController::class, 'count']);
Route::get('/write-request/done', [WriteRequestController::class, 'indexDone']);
Route::get('/write-request/todo', [WriteRequestController::class, 'indexTodo']);
Route::get('/write-request/todo-top', [WriteRequestController::class, 'indexTodoTop']);
Route::post('/write-request', [WriteRequestController::class, 'store'])->middleware('mwauth');
Route::delete('/write-request/{id}', [WriteRequestController::class, 'destroy'])->middleware('mwauth');

Route::get('/internal/profiles/{userId}', [UserProfileController::class, 'show'])->middleware('internal');
