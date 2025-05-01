<?php

namespace App\Http\Controllers;

use App\Jobs\CommonReportJob;
use App\Models\CommonReport;
use App\Models\CommonReportItem;
use Illuminate\Http\Request;

class CommonReportController extends MyController
{
    public function index()
    {
        return CommonReport::orderByDesc('id')->paginate(15);
    }

    public function show(string $id): CommonReport
    {
        return CommonReport::findOrFail($id);
    }

    public function store(Request $request)
    {
        $err = $this->shouldCreatable();
        if ($err !== false) {
            return $err;
        }

        $names = array_filter($request->input('names', []));
        if (! is_array($names) || count($names) < 2) {
            return $this->newHTTPError(422, '비교 대상을 2개 이상 입력해 주세요.');
        }

        $userId = $this->getMe()['avatar']['id'];

        CommonReport::create([
            'user_id' => $userId,
            'state' => 0,
        ]);

        foreach ($names as $name) {
            CommonReportItem::create([
                'report_id' => $report->id,
                'name' => $name,
            ]);
        }

        CommonReportJob::dispatch($report->id);
    }

    public function destroy($id)
    {
        $report = CommonReport::find($id);
        if (! $report) {
            return $this->newHTTPError(404, '해당 리포트가 없습니다.');
        }

        if ($err = $this->shouldDeletable($report->user_id)) {
            return $err;
        }

        $report->delete();

        return ['status' => 'ok'];
    }

    public function rerun(int $id)
    {
        $report = CommonReport::findOrFail($id);

        if ($err = $this->shouldRunnable($report->user_id)) {
            return $err;
        }

        $report->update(['state' => 0]);
        CommonReportJob::dispatch($report->id);
    }

    public function clone(int $id)
    {
        $original = CommonReport::findOrFail($id);

        if ($err = $this->shouldCreatable()) {
            return $err;
        }

        $userId = $this->getMe()['avatar']['id'];

        $clone = CommonReport::create([
            'user_id' => $userId,
            'state' => 0,
        ]);

        $items = CommonReportItem::where('report_id', $original->id)->get();

        foreach ($items as $item) {
            CommonReportItem::create([
                'report_id' => $clone->id,
                'name' => $item->name,
            ]);
        }

        CommonReportJob::dispatch($clone->id);
    }
}
