<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('stat_daily_ga', function (Blueprint $table) {
            $table->id();
            $table->date('timeslot')->comment('GA property date');
            $table->unsignedBigInteger('active_users')->default(0);
            $table->unsignedBigInteger('screen_page_views')->default(0);
            $table->unsignedBigInteger('sessions')->default(0);

            $table->unique('timeslot', 'stat_daily_ga_timeslot_unique');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('stat_daily_ga');
    }
};
