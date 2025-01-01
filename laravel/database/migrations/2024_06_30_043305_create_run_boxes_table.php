<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::create('run_boxes', function (Blueprint $table) {
            $table->id(); // identifier for backend
            $table->string('hash', length: 32); // identifier for frontend
            $table->unsignedMediumInteger('pageid');
            $table->smallInteger('step');
            $table->string('lang', length: 16);
            $table->text('files');
            $table->text('logs');
            $table->string('time', length: 16);
            $table->string('cpu', length: 8);
            $table->string('mem', length: 8);
            $table->string('error', length: 32);
            $table->timestamps();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('run_boxes');
    }
};
