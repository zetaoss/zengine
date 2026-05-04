<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (Schema::hasTable('doc_tasks')) {
            return;
        }

        Schema::create('doc_tasks', function (Blueprint $table) {
            $table->charset = 'utf8mb4';
            $table->collation = 'utf8mb4_general_ci';

            $table->id();
            $table->unsignedBigInteger('user_id');
            $table->string('user_name');
            $table->string('title');
            $table->string('request_type');
            $table->longText('content')->nullable();
            $table->string('status');
            $table->unsignedInteger('attempts')->default(0);
            $table->unsignedInteger('error_count')->default(0);
            $table->unsignedInteger('skip_count')->default(0);
            $table->text('last_error')->nullable();
            $table->timestamps();

            $table->index(['status', 'id']);
            $table->index(['request_type', 'status', 'id']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('doc_tasks');
    }
};
