outputPath = System.getenv('DTC_OUTPUT_PATH') ?: 'build/docs'
inputPath  = 'project/req42-tracer/docs'

inputFiles = [
    [file: 'user-guide/pdf.adoc',                     formats: ['pdf']],
    [file: 'user-guide/index.adoc',                   formats: ['html']],
    [file: 'user-guide/getting-started.adoc',         formats: ['html']],
    [file: 'user-guide/command-reference.adoc',       formats: ['html']],
    [file: 'user-guide/configuration-reference.adoc', formats: ['html']],
    [file: 'user-guide/block-attribute-reference.adoc', formats: ['html']],
    [file: 'user-guide/workflow-guide.adoc',          formats: ['html']],
    [file: 'arc42/arc42.adoc',                        formats: ['html', 'pdf']],
]

imageDirs = ["${inputPath}/images"]
