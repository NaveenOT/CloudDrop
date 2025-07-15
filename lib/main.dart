import 'dart:convert';
import 'dart:math';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:file_selector/file_selector.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:process_run/shell.dart';
import 'dart:async';
import 'package:http/http.dart' as http;
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Cloud Drop',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
      ),
      home: const MyHomePage(),
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key});

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}
class Meta{
  final String name;
  final String type;
  final String time;
  Meta({required this.name, required this.type, required this.time}); // is this constructor
  Meta.fromJSON(Map<String, dynamic> json): name = json['name'], type= json['type'], time=json['time'];
}
class _MyHomePageState extends State<MyHomePage> {
  String? savePath;
  String? tunnelURL;
  Process? goServer;
  Process? cf;
  Future<void> getDirPath() async{
    final String? path = await getDirectoryPath();
    setState(() {
      savePath = path;
      print(savePath);
    });
  }
  Future<List<Meta>> fetchFiles(String tunnel) async{
      final res = await http.get(Uri.parse('$tunnel/files'));
      final List<dynamic> filejson = jsonDecode(res.body);
      return filejson.map((e)=>Meta.fromJSON(e)).toList(); // IMPORTANT
  }
  Future<void> uploadFiles(String tunnel) async{
    final XFile? file = await openFile();
    final req = await http.MultipartRequest("POST", Uri.parse('$tunnel/upload'));
    req.files.add(await http.MultipartFile.fromPath('file', file!.path));
    final res = await req.send();
    if(res.statusCode == 200){
      print("Uploaded Successfully");
    }else{
      print("Error Sending File: $res.statusCode");
    }
  }
  Future<void> downloadFile(String tunnel, String filename) async{
    final res = await http.get(Uri.parse('$tunnel/download/$filename'));
    if(res.statusCode == 200){
      final destdir = await getDownloadsDirectory();
      final String dest = '${destdir!.path}/$filename';
      final bytes = res.bodyBytes; // actual content of the file in binary
      final file = File(dest);
      await file.writeAsBytes(bytes);
    }else{
      print('Error in Downloading: $res.statusCode');
    }
  }
  
  Future<void> createServer() async{
    final String execDir = File(Platform.resolvedExecutable).parent.path;
    final String serverDir = "$execDir/server/main.exe"; 
    goServer = await Process.start(serverDir, [savePath!],workingDirectory: Directory(serverDir).parent.path, runInShell: true);
  }
  Future<void> cfTunnel() async {
  cf = await Process.start('cloudflared', ['tunnel', '--url', 'http://localhost:8080'], runInShell: true);

  final completer = Completer<void>();

  cf?.stderr.transform(SystemEncoding().decoder).listen((line) {
    print('[CF] $line');
    final regex = RegExp(r'https:\/\/[a-zA-Z0-9\-]+\.trycloudflare\.com');
    final match = regex.firstMatch(line);

    if (match != null && !completer.isCompleted) {
      setState(() {
        tunnelURL = match.group(0);
        print('Tunnel URL: $tunnelURL');
      });
      completer.complete(); // this signals that we got the URL
    }
  });
  await completer.future; // wait until tunnelURL is received
}


  Future<void> startServer() async{
    if(savePath == null) return;
    print("Starting Server");
    await createServer();
    print("Server Started");
    await Future.delayed(Duration(seconds: 3));
    print("Tunneling Started");
    await cfTunnel();
    print("tunneling completed");
  }
  @override
  void dispose(){
    goServer?.kill();
    cf?.kill();
    super.dispose();
  }
  Future<List<Meta>>? metaList;
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          Text('Choose File to start'),
          ElevatedButton(onPressed: getDirPath, child: Text('Select File')),
          Text('$savePath'),
          Text('Create Server & Start'),
          ElevatedButton(onPressed: startServer, child: Text('Create Server')),
          const SizedBox(height: 20),
          if(tunnelURL != null)...[ // append a widget with condition
            QrImageView(data: tunnelURL!, size: 100),
          ],
          const SizedBox(height: 10),
          ElevatedButton(
              onPressed: () {
                setState(() {
                  metaList = fetchFiles(tunnelURL!);
                });
              },
              child: const Text("📂 Fetch Files"),
            ),
          const Text("📄 Available Files:", style: TextStyle(fontWeight: FontWeight.bold)),
          if(metaList != null)...[
            Expanded(child: FutureBuilder(future: metaList,
            builder:(context, snapshot){
              if(snapshot.connectionState == ConnectionState.waiting){
                return CircularProgressIndicator();
              }else if(snapshot.hasError){
                return Text("Error: $snapshot.error");
              }else if(!snapshot.hasData || snapshot.data!.isEmpty){
                  return Text('No Files');
              }
              final files = snapshot.data;
              return ListView.builder(//what does it do
                itemCount: files!.length,
                itemBuilder: (context, index){//learn what it does
                  final file = files[index];
                  return ListTile(
                    title: Text(file.name),
                    subtitle: Text("Type: ${file.type}\nTime: ${file.time}"),
                    trailing: IconButton(onPressed: ()=>downloadFile(tunnelURL!, file.name), icon: Icon(Icons.download)),
                  );
                },

              );
            }
             )
             )
          ],
        ],
      ),
    );
  }
}
