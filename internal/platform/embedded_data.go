package platform

import _ "embed"

// embeddedData contains a minimal fallback dataset
// In production, this would be loaded from S3 or local file
var embeddedData = []byte(`{"pmid":"12345678","title":"Ibuprofen and its clinical use in pain management","abstract":"Ibuprofen is a nonsteroidal anti-inflammatory drug commonly used for pain relief and inflammation reduction.","authors":["Smith J","Lee K"],"journal":"J Clin Pharm","pub_year":2020,"mesh_terms":["Ibuprofen","Anti-Inflammatory Agents"],"doi":"10.1000/jcp.2020.1234"}
{"pmid":"12345679","title":"Comparative study of ibuprofen and acetaminophen","abstract":"This study compares the efficacy of ibuprofen versus acetaminophen in treating postoperative pain.","authors":["Johnson A","Brown M"],"journal":"Pain Medicine","pub_year":2021,"mesh_terms":["Ibuprofen","Acetaminophen","Pain Management"],"doi":"10.1000/pm.2021.5678"}
`)

