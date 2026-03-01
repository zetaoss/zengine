<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('cf_analytics_daily', function (Blueprint $table) {
            $table->id();
            $table->date('timeslot')->comment('UTC date');
            $table->string('name', 64);
            $table->text('value')->comment('JSON text');

            $table->unique(['timeslot', 'name'], 'cf_analytics_daily_timeslot_name_unique');
            $table->index('timeslot', 'cf_analytics_daily_timeslot_index');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('cf_analytics_daily');
    }
};
