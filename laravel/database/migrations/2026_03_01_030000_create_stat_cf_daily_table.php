<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (Schema::hasTable('stat_cf_daily')) {
            return;
        }

        Schema::create('stat_cf_daily', function (Blueprint $table) {
            $table->id();
            $table->date('timeslot')->comment('UTC date');
            $table->string('name', 64);
            $table->text('value')->comment('JSON text');

            $table->unique(['timeslot', 'name'], 'stat_cf_daily_timeslot_name_unique');
            $table->index('timeslot', 'stat_cf_daily_timeslot_index');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('stat_cf_daily');
    }
};
