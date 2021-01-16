package com.example;

import net.sourceforge.argparse4j.ArgumentParsers;
import net.sourceforge.argparse4j.impl.Arguments;
import net.sourceforge.argparse4j.inf.ArgumentParser;
import net.sourceforge.argparse4j.inf.ArgumentParserException;
import net.sourceforge.argparse4j.inf.MutuallyExclusiveGroup;
import net.sourceforge.argparse4j.inf.Namespace;

import java.io.*;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.IntStream;
import java.util.stream.Collectors;

/**
 * Hello world!
 *
 */
public class App
{
    private static Namespace parseArgs(String[] args) throws ArgumentParserException {
        ArgumentParser parser = ArgumentParsers.newFor("grepy").build()
                .defaultHelp(true)
                .description("looks for regex in file(s)");
        parser.addArgument("regex")
                .help("Regex to match in files");
        parser.addArgument("files")
                .nargs("*")
                .type(Arguments.fileType().acceptSystemIn().verifyCanRead())
                .setDefault("-")
                .help("Files to look in, if not provided will read from stdin");
        MutuallyExclusiveGroup fmt = parser.addMutuallyExclusiveGroup();
        fmt.addArgument("-c", "--color")
                .dest("format")
                .action(Arguments.storeConst())
                .setConst(color)
                .help("Highlights the matches in color");
        fmt.addArgument("-u", "--underline")
                .dest("format")
                .action(Arguments.storeConst())
                .setConst(underline)
                .help("Highlight the matches with '^' under the line");
        fmt.addArgument("-m", "--machine")
                .dest("format")
                .action(Arguments.storeConst())
                .setConst(machine)
                .help("Print in machine format [file name]:[line number]:[match text]");
        parser.setDefault("format", machine);
        return parser.parseArgs(args);
    }

    public static void main( String[] args ) {
        Namespace params;
        try {
            params = parseArgs(args);
        } catch (ArgumentParserException e) {
            return;
        }
        Pattern pattern = Pattern.compile(params.getString("regex"));
        Formatter formatter = params.get("format");
        for(Object o: params.getList("files")) {
            File f = (File)o;
            BufferedReader reader = null;
            try {
                reader = new BufferedReader(new FileReader(f));
            } catch (FileNotFoundException e) {
                // should never happen as argparse already checked for this
                e.printStackTrace();
            }
            String line;
            int line_num = 0;
            try {
                while ((line = reader.readLine()) != null) {
                    line_num += 1;
                    Matcher matcher = pattern.matcher(line);
                    if (matcher.find()) {
                        formatter.format(f.getName(), line_num, line, matcher);
                    }
                }
            } catch (IOException e) {
                e.printStackTrace();
            }
        }
    }

    private interface Formatter {
        void format(String filename, int line_num, String line, Matcher matcher);
    }

    private static final Formatter color = (filename, line_num, line, matcher) -> {
        StringBuilder builder = new StringBuilder();
        builder.append(filename).append(" (").append(line_num).append(") ");
        int last_start = 0;
        do {
            builder.append(line.substring(last_start, matcher.start()))
                    .append("\033[32m")
                    .append(line.substring(matcher.start(), matcher.end()))
                    .append("\33[39m");
            last_start = matcher.end();
        } while (matcher.find());
        System.out.println(builder.toString());
    };

    private static String repeatStr(String str, int times) {
        return IntStream.range(0, times).mapToObj(i -> str).collect(Collectors.joining(""));
    }

    private static final Formatter underline = (filename, line_num, line, matcher) -> {
        StringBuilder builder = new StringBuilder();

        // first print the line
        builder.append(filename).append(" (").append(line_num).append(") ");
        int prefix_len = builder.length();
        builder.append(line);
        System.out.println(builder.toString());

        // now print the highlights
        builder = new StringBuilder(builder.length());
        builder.append(repeatStr(" ", prefix_len));
        int last_start = 0;
        do {
            builder.append(repeatStr(" ", matcher.start() - last_start))
                    .append(repeatStr("^", matcher.end() - matcher.start()));
            last_start = matcher.end();
        } while (matcher.find());
        System.out.println(builder.toString());
    };

    private static final Formatter machine = (filename, line_num, line, matcher) -> {
        System.out.println(String.format("%s:%d:%s", filename, line_num, matcher.group()));
    };
}
