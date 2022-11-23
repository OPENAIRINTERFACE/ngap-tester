BEGIN{n=4000}
{
  if (($0 ~/208950000000128/) && ($0 ~/defaultSingleNssais/)) {
    print $0
    for (idx = 0 ; idx < n; idx ++) {
      mLine = $0
      new_imsi = sprintf("2089500%08d", 130 + idx)
      gsub("208950000000128", new_imsi, mLine)
      print mLine
    }
  } else if (($0 ~/208950000000130/) && ($0 ~/5G_AKA/)) {
    print $0
    for (idx = 0 ; idx < n; idx ++) {
      mLine = $0
      new_imsi = sprintf("2089500%08d", 132 + idx)
      gsub("208950000000130", new_imsi, mLine)
      print mLine
    }
  } else {
    print $0
  }
}
END{}
