package main

import (
    "net/http"
    "html/template"
    "github.com/gorilla/mux"
    "github.com/jung-kurt/gofpdf"
    "github.com/nfnt/resize"
    "io"
    "os"
    "github.com/unidoc/unioffice/document"
    "image"
    "image/jpeg"
    "bytes"
)

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/", initialHandler).Methods("GET")
    r.HandleFunc("/convert", conversionHandler).Methods("POST")
   // r.HandleFunc("/download", downloadHandler).Methods("GET")

    http.Handle("/", r)
    http.ListenAndServe(":8080", nil)
}

func initialHandler(w http.ResponseWriter, r *http.Request) {
    temp, err := template.New("init").Parse(initHTML)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    temp.Execute(w, nil)
}

const initHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>File Converter</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #fff;
            padding: 20px;
            border-radius: 5px;
            box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
        }
        h1 {
            color: #333;
            text-align: center;
        }
        form {
            margin-top: 20px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
        }
        input[type="radio"] {
            margin-right: 5px;
        }
        input[type="file"] {
            margin-top: 5px;
        }
        button[type="submit"] {
            background-color: #007bff;
            color: #fff;
            padding: 10px 20px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
        }
        button[type="submit"]:hover {
            background-color: #0056b3;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Select File Type to Convert</h1>
        <form action="/convert" method="post" enctype="multipart/form-data">
            <label>
                <input type="radio" name="fileType" value="text"> Convert Text to PDF
            </label>
            <label>
                <input type="radio" name="fileType" value="image"> Convert Image to JPEG
            </label>
            <label>
                <input type="radio" name="fileType" value="docx"> Convert DOCX to PDF
            </label>
            <label>
                <input type="radio" name="fileType" value="doc"> Convert DOC to PDF
            </label>
            <input type="file" name="file" required>
            <button type="submit">Convert</button>
        </form>
    </div>
</body>
</html>
`

func conversionHandler(w http.ResponseWriter, r *http.Request) {
    typeOfFile := r.FormValue("typeOfFile")

    file, _, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Error uploading the file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    if typeOfFile == "text" {
        convertToPDF(file, w, r)
    } else if typeOfFile == "image" {
        convertToImage(file, w, r)
    } else if typeOfFile == "docx" {
        convertToPDFFromDocx(file, w, r)
    } else {
        http.Error(w, "Invalid file type", http.StatusBadRequest)
    }
}

func convertToPDF(file io.Reader, w http.ResponseWriter, r *http.Request) {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    pdf.SetFont("Arial", "", 16)

    text, err := io.ReadAll(file)
    if err != nil {
        http.Error(w, "Error reading the file", http.StatusInternalServerError)
        return
    }

    pdf.MultiCell(0, 10, string(text), "", "", false)

    pdfFileName := "output.pdf"
    pdf.OutputFileAndClose(pdfFileName)

    w.Header().Set("Content-Disposition", "attachment; filename="+pdfFileName)
    w.Header().Set("Content-Type", "application/pdf")
    http.ServeFile(w, r, pdfFileName)
}

func convertToImage(file io.Reader, w http.ResponseWriter, r *http.Request) {
    img, _, err := image.Decode(file)
    if err != nil {
        http.Error(w, "Error decoding image", http.StatusInternalServerError)
        return
    }

    img = resize.Resize(200, 0, img, resize.Lanczos3)
    out, err := os.Create("output.jpg")
    if err != nil {
        http.Error(w, "Error creating image output", http.StatusInternalServerError)
        return
    }
    defer out.Close()

    jpegOptions := &jpeg.Options{Quality: 100}
    err = jpeg.Encode(out, img, jpegOptions)
    if err != nil {
        http.Error(w, "Error encoding image", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Disposition", "attachment; filename=output.jpg")
    w.Header().Set("Content-Type", "image/jpeg")
    http.ServeFile(w, r, "output.jpg")
}

func convertToPDFFromDocx(file io.Reader, w http.ResponseWriter, r *http.Request) {
    data, err := io.ReadAll(file)
    if err != nil {
        http.Error(w, "Error reading the DOCX file", http.StatusInternalServerError)
        return
    }

    reader := bytes.NewReader(data)
    doc, err := document.Read(reader, int64(len(data)))
    if err != nil {
        http.Error(w, "Error reading the DOCX file", http.StatusInternalServerError)
        return
    }

    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()

    for _, para := range doc.Paragraphs() {
        text := extractTextFromRuns(para.Runs())
        pdf.SetFont("Arial", "", 16)
        pdf.MultiCell(0, 10, text, "", "", false)
    }

    pdfFileName := "output.pdf"
    pdf.OutputFileAndClose(pdfFileName)

    w.Header().Set("Content-Disposition", "attachment; filename=output.pdf")
    w.Header().Set("Content-Type", "application/pdf")
    http.ServeFile(w, r, pdfFileName)
}



func extractTextFromRuns(runs []document.Run) string {
    var text string
    for _, run := range runs {
        text += run.Text()
    }
    return text
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
    // Implement the download handler as needed
}
