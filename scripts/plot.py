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
        "run_time" : benchmark_result['fullRunStats']["max"], 
        # "run_time" : benchmark_result['lazyTaskStats']["max"], 
        # "unpackStats" : benchmark_result['unpackStats']["max"],
        "full_run_stats": benchmark_result['fullRunStats']['max']
     }


# plot 
ROOT.gStyle.SetOptStat(0)
# ROOT.gStyle.SetLegendTextSize(0.04)
for image, snapshotters in benchmark_result_cleaned.items():
    canvas = ROOT.TCanvas(image, image, 500, 500)
    stack = ROOT.THStack("ts", image)
    stack_items = ["pull_time", "create_time", "run_time"]
    fill_styles = [0, 3005, 3001]
    fill_colors = [ROOT.kGray, ROOT.kGray+2, ROOT.kBlack]
    legend = ROOT.TLegend(0.7,0.7,0.8,0.898)

    for item_idx, item in enumerate(stack_items):
        for snapshotter_idx, snapshotter in enumerate(snapshotters):
            histogram = ROOT.TH1F(f"{image}-{snapshotter}-{item}", f"{image};;Time [s];", len(snapshotters), -0.5, len(snapshotters) - 0.5)
            histogram.SetBinContent(snapshotter_idx + 1, benchmark_result_cleaned[image][snapshotter][item])
            histogram.GetXaxis().SetTickLength(0)
            histogram.GetYaxis().SetTitleOffset(1.7)
            histogram.GetXaxis().SetLabelSize(0.06)
            histogram.SetMinimum(0)
            histogram.SetLineColor(1)
            histogram.SetFillStyle(fill_styles[item_idx])
            histogram.SetFillColor(fill_colors[item_idx])

            if item_idx == 0 and snapshotter_idx == 0:
                for s_idx, s_name in enumerate(snapshotters):
                    histogram.GetXaxis().SetBinLabel(s_idx + 1, s_name)

            if snapshotter_idx == 0:
                label = " ".join([w.capitalize() for w in item.split("_")])
                legend.AddEntry(histogram, label, "f")

            stack.Add(histogram)
    
    canvas.Draw()
    stack.Draw("hist")
    stack.GetXaxis().SetTickLength(0)
    stack.GetXaxis().SetLabelSize(0.06)
    stack.GetXaxis().SetTickLength(0)
    # ROOT.gPad.SetLeftMargin(0.15)
    stack.GetYaxis().SetTitleOffset(1.22)
    stack.GetYaxis().SetTitle("Time [s]")
    legend.SetTextSize(0.03)
    legend.SetBorderSize(0)
    legend.Draw()
    canvas.SaveAs(f"output/plots/{image}.png")
