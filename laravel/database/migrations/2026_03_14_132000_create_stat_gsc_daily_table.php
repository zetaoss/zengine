<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('stat_gsc_daily', function (Blueprint $table) {
            $table->id();
            $table->date('timeslot')->comment('Search Console property date');
            $table->unsignedBigInteger('clicks')->default(0);
            $table->unsignedBigInteger('impressions')->default(0);
            $table->decimal('ctr', 9, 4)->default(0);
            $table->decimal('position', 9, 4)->default(0);

            $table->unique('timeslot', 'stat_gsc_daily_timeslot_unique');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('stat_gsc_daily');
    }
};
