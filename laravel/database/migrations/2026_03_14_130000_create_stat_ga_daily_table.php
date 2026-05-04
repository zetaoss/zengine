<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (Schema::hasTable('stat_ga_daily')) {
            return;
        }

        Schema::create('stat_ga_daily', function (Blueprint $table) {
            $table->id();
            $table->date('timeslot')->comment('GA property date');
            $table->unsignedBigInteger('active_users')->default(0);
            $table->unsignedBigInteger('screen_page_views')->default(0);
            $table->unsignedBigInteger('sessions')->default(0);

            $table->unique('timeslot', 'stat_ga_daily_timeslot_unique');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('stat_ga_daily');
    }
};
