#!/usr/bin/env python3

import json
import ROOT
import os
from collections import defaultdict

benchmark_results = dict()
commit = ""

os.makedirs("output/plots", exist_ok=True)
with open("output/results.json", 'r') as f:
    data = json.loads(f.read())
    commit = data["commit"]
    benchmark_results = data["benchmarkTests"]

benchmark_result_cleaned = defaultdict(dict)

# reformat data
"""
    For example:
        {
            "python:3.9": {
                "overlayfs": {
                    "pull_time": 1.0
                    "create_time": 1.5
                    "run_time": 2.0
                }
            }
        }
"""
for benchmark_result in benchmark_results:
    snapshotter, image = benchmark_result['testName'].split(" ") 
    benchmark_result_cleaned[image][snapshotter] = { 
        "pull_time" : benchmark_result['pullStats']["max"],
        "create_time" : benchmark_result['createStats']["max"], 
        "run_time" : benchmark_result['lazyTaskStats']["max"], 
        # "unpackStats" : benchmark_result['unpackStats']["max"],
        # "full_run_stats": benchmark_result['fullRunStats']['max']
     }


# plot 

ROOT.gStyle.SetOptStat(0)
for image, snapshotters in benchmark_result_cleaned.items():
    canvas = ROOT.TCanvas(image, image, 500, 500)
    histogram = ROOT.TH1F(image, f"{image};;Time [s];", len(snapshotters), -0.5, len(snapshotters) - 0.5)

    for i, snapshotter in enumerate(snapshotters):
        histogram.SetBinContent(i + 1, benchmark_result_cleaned[image][snapshotter]["full_run_stats"])
        histogram.GetXaxis().SetBinLabel(i + 1, snapshotter)
        histogram.GetXaxis().SetLabelSize(0.06)
        histogram.GetXaxis().SetTickLength(0)
        ROOT.gPad.SetLeftMargin(0.15)
        histogram.GetYaxis().SetTitleOffset(1.7)
        histogram.SetMinimum(0)
        
    canvas.Draw()
    histogram.Draw()
    canvas.SaveAs(f"output/plots/{image}.png")
