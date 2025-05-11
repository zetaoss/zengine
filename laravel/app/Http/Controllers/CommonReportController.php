<?php

namespace App\Http\Controllers;

use App\Models\CommonReport;
use App\Services\CommonReportService;
use Illuminate\Http\Request;

class CommonReportController extends MyController
{
    protected CommonReportService $reportService;

    public function __construct(CommonReportService $reportService)
    {
        $this->reportService = $reportService;
    }

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
        $report = $this->reportService->create($names, $userId);

        return $report;
    }

    public function rerun(int $id)
    {
        $report = CommonReport::findOrFail($id);
        if ($err = $this->shouldRunnable($report->user_id)) {
            return $err;
        }

        $this->reportService->rerunReport($report);
    }

    public function clone(int $id)
    {
        $original = CommonReport::findOrFail($id);
        if ($err = $this->shouldCreatable()) {
            return $err;
        }

        $userId = $this->getMe()['avatar']['id'];
        $this->reportService->cloneReport($original, $userId);
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

        $this->reportService->deleteReport($report);

        return ['status' => 'ok'];
    }
}
