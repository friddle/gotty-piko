package com.gottyp.android;

import android.Manifest;
import android.app.Activity;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.net.Uri;
import android.os.Bundle;
import android.provider.Settings;
import android.util.Log;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.Switch;
import android.widget.TextView;
import android.widget.Toast;

import androidx.appcompat.app.AppCompatActivity;
import androidx.core.app.ActivityCompat;
import androidx.core.content.ContextCompat;

import java.io.BufferedReader;
import java.io.DataOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;

// Import Go mobile bindings
import gottyp.Gottyp;
import gottyp.ServiceWrapper;

public class MainActivity extends AppCompatActivity {
    private static final String TAG = "GottypAndroid";
    private static final int PERMISSION_REQUEST_CODE = 1001;
    
    private EditText remoteAddrEdit;
    private Switch gottypSwitch;
    private Switch debugSwitch;
    private TextView statusText;
    private TextView rootStatusText;
    private Button settingsButton;
    
    private boolean hasRoot = false;
    private String rootCommand = ""; // 存储可用的root命令
    private boolean isGottypRunning = false;
    private boolean isDebugEnabled = false;
    
    // Go mobile service instance
    private ServiceWrapper gottypService;
    
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        
        initViews();
        initGottypService();
        extractBinary();
        checkRootPermission();
        checkPermissions();
    }
    
    private void initViews() {
        remoteAddrEdit = findViewById(R.id.remote_addr_edit);
        gottypSwitch = findViewById(R.id.gottyp_switch);
        debugSwitch = findViewById(R.id.debug_switch);
        statusText = findViewById(R.id.status_text);
        rootStatusText = findViewById(R.id.root_status_text);
        settingsButton = findViewById(R.id.settings_button);
        
        // Set default remote address
        remoteAddrEdit.setText("https://remote-coide-test.code27.cn");
        
        // Set listeners
        gottypSwitch.setOnCheckedChangeListener((buttonView, isChecked) -> {
            if (isChecked) {
                startGottypMode();
            } else {
                stopGottypMode();
            }
        });
        
        debugSwitch.setOnCheckedChangeListener((buttonView, isChecked) -> {
            if (isChecked) {
                enableDebugMode();
            } else {
                disableDebugMode();
            }
        });
        
        settingsButton.setOnClickListener(v -> openDeveloperOptions());
        
        updateStatus();
    }
    
    private void initGottypService() {
        try {
            // Initialize Go mobile service
            gottypService = Gottyp.getService();
            Log.i(TAG, "GottypService initialized successfully");
        } catch (Exception e) {
            Log.e(TAG, "Failed to initialize GottypService", e);
            Toast.makeText(this, "Failed to initialize GottypService: " + e.getMessage(), Toast.LENGTH_SHORT).show();
        }
    }
    
    private void extractBinary() {
        try {
            // Get application internal storage directory
            File filesDir = getFilesDir();
            File gottypBinary = new File(filesDir, "gottyp");
            
            // If binary file already exists and is up to date, skip extraction
            if (gottypBinary.exists()) {
                Log.i(TAG, "Gottyp binary file already exists: " + gottypBinary.getAbsolutePath());
                return;
            }
            
            // Copy binary file from assets
            Log.i(TAG, "Extracting gottyp binary file from assets...");
            InputStream inputStream = getAssets().open("gottyp");
            FileOutputStream outputStream = new FileOutputStream(gottypBinary);
            
            byte[] buffer = new byte[8192];
            int bytesRead;
            while ((bytesRead = inputStream.read(buffer)) != -1) {
                outputStream.write(buffer, 0, bytesRead);
            }
            
            inputStream.close();
            outputStream.close();
            
            // Set executable permissions
            gottypBinary.setExecutable(true);
            
            Log.i(TAG, "Gottyp binary file extracted successfully: " + gottypBinary.getAbsolutePath());
            
        } catch (IOException e) {
            Log.e(TAG, "Failed to extract gottyp binary file", e);
            Toast.makeText(this, "Failed to extract binary file: " + e.getMessage(), Toast.LENGTH_SHORT).show();
        }
    }
    
    private void checkRootPermission() {
        // Try multiple root commands including suks
        String[] rootCommands = {"suks", "su", "sudo", "pksu"};
        hasRoot = false;
        rootCommand = "";
        
        for (String cmd : rootCommands) {
            if (tryRootCommand(cmd)) {
                hasRoot = true;
                rootCommand = cmd;
                Log.i(TAG, "Found available root command: " + cmd);
                break;
            }
        }
        
        runOnUiThread(() -> {
            if (hasRoot) {
                rootStatusText.setText("✅ Root Permission: Granted (" + rootCommand + ")");
                rootStatusText.setTextColor(ContextCompat.getColor(this, android.R.color.holo_green_dark));
            } else {
                rootStatusText.setText("❌ Root Permission: Not Granted - Click to retry");
                rootStatusText.setTextColor(ContextCompat.getColor(this, android.R.color.holo_red_dark));
                // 添加点击重试功能
                rootStatusText.setOnClickListener(v -> {
                    rootStatusText.setText("🔄 Checking root permission...");
                    rootStatusText.setTextColor(ContextCompat.getColor(this, android.R.color.holo_orange_dark));
                    new Thread(this::checkRootPermission).start();
                });
            }
        });
    }
    
    private boolean tryRootCommand(String command) {
        try {
            Process process;
            DataOutputStream os;
            
            // Use different test methods based on different commands
            if ("sudo".equals(command)) {
                // sudo usually requires password, we try passwordless execution
                process = Runtime.getRuntime().exec(command);
                os = new DataOutputStream(process.getOutputStream());
                os.writeBytes("sudo -n id\n");
                os.flush();
            } else if ("pksu".equals(command)) {
                // pksu may need specific parameters
                process = Runtime.getRuntime().exec(command);
                os = new DataOutputStream(process.getOutputStream());
                os.writeBytes("pksu -c id\n");
                os.flush();
            } else if ("suks".equals(command)) {
                // suks command test - execute suks root id directly
                Log.i(TAG, "Testing suks command directly...");
                process = Runtime.getRuntime().exec("suks root id");
                os = null; // No need for DataOutputStream for direct execution
            } else {
                // Default su command
                process = Runtime.getRuntime().exec(command);
                os = new DataOutputStream(process.getOutputStream());
                os.writeBytes("id\n");
                os.flush();
            }
            
            if (os != null) {
                os.writeBytes("exit\n");
                os.flush();
            }
            
            int exitCode = process.waitFor();
            boolean success = (exitCode == 0);
            
            if (success) {
                Log.i(TAG, "Root command " + command + " test successful");
            } else {
                Log.d(TAG, "Root command " + command + " test failed, exit code: " + exitCode);
            }
            
            return success;
            
        } catch (Exception e) {
            Log.d(TAG, "Root command " + command + " not available: " + e.getMessage());
            return false;
        }
    }
    
    private void checkPermissions() {
        if (ContextCompat.checkSelfPermission(this, Manifest.permission.WRITE_EXTERNAL_STORAGE) 
                != PackageManager.PERMISSION_GRANTED) {
            ActivityCompat.requestPermissions(this, 
                new String[]{Manifest.permission.WRITE_EXTERNAL_STORAGE}, 
                PERMISSION_REQUEST_CODE);
        }
    }
    
    private void startGottypMode() {
        String remoteAddr = remoteAddrEdit.getText().toString().trim();
        if (remoteAddr.isEmpty()) {
            Toast.makeText(this, "Please enter remote address", Toast.LENGTH_SHORT).show();
            gottypSwitch.setChecked(false);
            return;
        }
        
        // Start gottyp service
        startGottypService();
    }
    
    private void stopGottypMode() {
        // Stop gottyp service
        stopGottypService();
    }
    
    private void enableDebugMode() {
        if (hasRoot) {
            enableRootDebug();
        } else {
            enableNonRootDebug();
        }
    }
    
    private void enableRootDebug() {
        try {
            Process process = Runtime.getRuntime().exec(rootCommand);
            DataOutputStream os = new DataOutputStream(process.getOutputStream());
            
            // Build commands based on different root commands
            String setPropCmd = buildSetPropCommand("setprop service.adb.tcp.port 5555");
            String persistCmd = buildSetPropCommand("setprop persist.adb.tcp.port 5555");
            String stopCmd = buildCommand("stop adbd");
            String startCmd = buildCommand("start adbd");
            
            // Set fixed port
            os.writeBytes(setPropCmd + "\n");
            os.flush();
            
            // Enable remote debugging
            os.writeBytes(persistCmd + "\n");
            os.flush();
            
            // Restart ADB service
            os.writeBytes(stopCmd + "\n");
            os.flush();
            os.writeBytes(startCmd + "\n");
            os.flush();
            
            os.writeBytes("exit\n");
            os.flush();
            
            int exitCode = process.waitFor();
            if (exitCode == 0) {
                isDebugEnabled = true;
                updateStatus();
                Toast.makeText(this, "Root debug mode enabled (" + rootCommand + ")", Toast.LENGTH_SHORT).show();
            } else {
                debugSwitch.setChecked(false);
                Toast.makeText(this, "Failed to enable root debug mode", Toast.LENGTH_SHORT).show();
            }
            
        } catch (Exception e) {
            Log.e(TAG, "Failed to enable root debug mode", e);
            debugSwitch.setChecked(false);
            Toast.makeText(this, "Failed to enable root debug mode: " + e.getMessage(), Toast.LENGTH_SHORT).show();
        }
    }
    
    private String buildCommand(String cmd) {
        if ("sudo".equals(rootCommand)) {
            return "sudo " + cmd;
        } else if ("pksu".equals(rootCommand)) {
            return "pksu -c \"" + cmd + "\"";
        } else if ("suks".equals(rootCommand)) {
            return "suks root " + cmd;
        } else {
            return cmd;
        }
    }
    
    private String buildSetPropCommand(String cmd) {
        return buildCommand(cmd);
    }
    
    private String getGottypBinaryPath() {
        File filesDir = getFilesDir();
        File gottypBinary = new File(filesDir, "gottyp");
        return gottypBinary.getAbsolutePath();
    }
    
    private void startGottypService() {
        String remoteAddr = remoteAddrEdit.getText().toString().trim();
        
        if (remoteAddr.isEmpty()) {
            remoteAddr = "https://remote-coide-test.code27.cn:8022";
        }
        
        if (hasRoot) {
            // Root mode: use binary execution
            startGottypServiceRoot(remoteAddr);
        } else {
            // Non-root mode: use Go mobile service
            startGottypServiceNonRoot(remoteAddr);
        }
    }
    
    private void startGottypServiceRoot(String remoteAddr) {
        try {
            String binaryPath = getGottypBinaryPath();
            
            // 详细日志输出
            Log.i(TAG, "=== 启动 Gottyp 服务 (Root 模式) ===");
            Log.i(TAG, "Root 命令: " + rootCommand);
            Log.i(TAG, "二进制路径: " + binaryPath);
            Log.i(TAG, "远程地址: " + remoteAddr);
            Log.i(TAG, "终端类型: sh");
            Log.i(TAG, "自动退出: false");
            
            // 构建启动命令
            String startCmd = buildCommand(binaryPath + " --name=sni --remote=" + remoteAddr + " --terminal=sh --auto-exit=false");
            Log.i(TAG, "执行命令: " + startCmd);
            
            Process process = Runtime.getRuntime().exec(rootCommand);
            DataOutputStream os = new DataOutputStream(process.getOutputStream());
            
            // 输出命令到日志
            os.writeBytes("echo '=== Gottyp Root 模式启动 ==='\n");
            os.writeBytes("echo 'Root 命令: " + rootCommand + "'\n");
            os.writeBytes("echo '二进制路径: " + binaryPath + "'\n");
            os.writeBytes("echo '远程地址: " + remoteAddr + "'\n");
            os.writeBytes("echo '执行命令: " + startCmd + "'\n");
            os.writeBytes("echo '开始启动服务...'\n");
            os.flush();
            
            // 执行启动命令
            os.writeBytes(startCmd + "\n");
            os.flush();
            
            // 等待并检查状态
            os.writeBytes("sleep 2\n");
            os.writeBytes("ps aux | grep gottyp\n");
            os.writeBytes("echo '服务启动完成'\n");
            os.flush();
            
            os.writeBytes("exit\n");
            os.flush();
            
            int exitCode = process.waitFor();
            
            Log.i(TAG, "Root 模式启动命令执行完成，退出码: " + exitCode);
            
            if (exitCode == 0) {
                isGottypRunning = true;
                updateStatus();
                Log.i(TAG, "✅ Gottyp 服务启动成功 (Root 模式)");
                Toast.makeText(this, "✅ Gottyp 服务启动成功 (Root 模式)", Toast.LENGTH_SHORT).show();
            } else {
                Log.e(TAG, "❌ Gottyp 服务启动失败 (Root 模式)，退出码: " + exitCode);
                Toast.makeText(this, "❌ Gottyp 服务启动失败 (Root 模式)", Toast.LENGTH_SHORT).show();
            }
            
        } catch (Exception e) {
            Log.e(TAG, "❌ 启动 Gottyp 服务时发生异常 (Root 模式)", e);
            Toast.makeText(this, "❌ 启动失败: " + e.getMessage(), Toast.LENGTH_SHORT).show();
        }
    }
    
    private void startGottypServiceNonRoot(String remoteAddr) {
        try {
            // 详细日志输出
            Log.i(TAG, "=== 启动 Gottyp 服务 (GoMobile 模式) ===");
            Log.i(TAG, "模式: GoMobile (非Root)");
            Log.i(TAG, "客户端名称: sni");
            Log.i(TAG, "远程地址: " + remoteAddr);
            Log.i(TAG, "终端类型: sh");
            Log.i(TAG, "密码: 无");
            
            // 使用gomobile调用Go服务
            if (gottypService != null) {
                Log.i(TAG, "GoMobile 服务已初始化，开始启动...");
                
                // 启动服务
                String error = gottypService.startService("sni", remoteAddr, "sh", "");
                
                if (error.isEmpty()) {
                    isGottypRunning = true;
                    updateStatus();
                    
                    // 获取详细状态信息
                    String detailedStatus = gottypService.getDetailedStatus();
                    Log.i(TAG, "✅ Gottyp 服务启动成功 (GoMobile 模式)");
                    Log.i(TAG, "详细状态: " + detailedStatus);
                    
                    // 获取本地端口
                    long localPort = gottypService.getLocalPort();
                    Log.i(TAG, "本地监听端口: " + localPort);
                    
                    Toast.makeText(this, "✅ Gottyp 服务启动成功 (GoMobile 模式)\n端口: " + localPort, Toast.LENGTH_LONG).show();
                } else {
                    Log.e(TAG, "❌ Gottyp 服务启动失败: " + error);
                    Toast.makeText(this, "❌ 启动失败: " + error, Toast.LENGTH_SHORT).show();
                }
            } else {
                Log.e(TAG, "❌ GoMobile 服务未初始化");
                Toast.makeText(this, "❌ GoMobile 服务未初始化", Toast.LENGTH_SHORT).show();
            }
            
        } catch (Exception e) {
            Log.e(TAG, "❌ 启动 Gottyp 服务时发生异常 (GoMobile 模式)", e);
            Toast.makeText(this, "❌ 启动失败: " + e.getMessage(), Toast.LENGTH_SHORT).show();
        }
    }
    
    private void stopGottypService() {
        if (hasRoot) {
            // Root mode: use process kill
            stopGottypServiceRoot();
        } else {
            // Non-root mode: use Go mobile service
            stopGottypServiceNonRoot();
        }
    }
    
    private void stopGottypServiceRoot() {
        try {
            Process process = Runtime.getRuntime().exec(rootCommand);
            DataOutputStream os = new DataOutputStream(process.getOutputStream());
            
            String stopCmd = buildCommand("pkill -f gottyp");
            os.writeBytes(stopCmd + "\n");
            os.flush();
            os.writeBytes("exit\n");
            os.flush();
            
            process.waitFor();
            
            isGottypRunning = false;
            updateStatus();
            Toast.makeText(this, "Gottyp service stopped (root mode)", Toast.LENGTH_SHORT).show();
            
        } catch (Exception e) {
            Log.e(TAG, "Failed to stop gottyp service (root mode)", e);
        }
    }
    
    private void stopGottypServiceNonRoot() {
        try {
            Log.i(TAG, "Stopping gottyp service (non-root mode)");
            
            // 使用gomobile调用Go服务停止
            if (gottypService != null) {
                // 停止服务
                String error = gottypService.stopService();
                
                if (error.isEmpty()) {
                    isGottypRunning = false;
                    updateStatus();
                    Toast.makeText(this, "Gottyp service stopped successfully (non-root mode)", Toast.LENGTH_SHORT).show();
                    Log.i(TAG, "Gottyp service stopped successfully via gomobile");
                } else {
                    Toast.makeText(this, "Failed to stop gottyp service: " + error, Toast.LENGTH_SHORT).show();
                    Log.e(TAG, "Failed to stop gottyp service: " + error);
                }
            } else {
                Toast.makeText(this, "GottypService not initialized", Toast.LENGTH_SHORT).show();
                Log.e(TAG, "GottypService not initialized");
            }
            
        } catch (Exception e) {
            Log.e(TAG, "Failed to stop gottyp service (non-root mode)", e);
            Toast.makeText(this, "Failed to stop gottyp service (non-root mode): " + e.getMessage(), Toast.LENGTH_SHORT).show();
        }
    }
    
    private void enableNonRootDebug() {
        // Prompt user to enable developer options and USB debugging
        Toast.makeText(this, "Please enable developer options and USB debugging", Toast.LENGTH_LONG).show();
        
        // Open developer options settings page
        Intent intent = new Intent(Settings.ACTION_APPLICATION_DEVELOPMENT_SETTINGS);
        if (intent.resolveActivity(getPackageManager()) != null) {
            startActivity(intent);
        } else {
            // If unable to open developer options directly, open app settings page
            intent = new Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS);
            intent.setData(Uri.fromParts("package", getPackageName(), null));
            startActivity(intent);
        }
        
        // Get debug port
        int debugPort = getDebugPort();
        if (debugPort > 0) {
            // Start ADB forward service
            startAdbForwardService(debugPort);
            isDebugEnabled = true;
            updateStatus();
            Toast.makeText(this, "Non-root debug mode enabled, port: " + debugPort, Toast.LENGTH_SHORT).show();
        } else {
            debugSwitch.setChecked(false);
            Toast.makeText(this, "Unable to get debug port", Toast.LENGTH_SHORT).show();
        }
    }
    
    private void disableDebugMode() {
        if (hasRoot) {
            try {
                Process process = Runtime.getRuntime().exec(rootCommand);
                DataOutputStream os = new DataOutputStream(process.getOutputStream());
                
                // Disable remote debugging
                String disableCmd1 = buildSetPropCommand("setprop service.adb.tcp.port -1");
                String disableCmd2 = buildSetPropCommand("setprop persist.adb.tcp.port -1");
                String stopCmd = buildCommand("stop adbd");
                String startCmd = buildCommand("start adbd");
                
                os.writeBytes(disableCmd1 + "\n");
                os.flush();
                os.writeBytes(disableCmd2 + "\n");
                os.flush();
                
                // Restart ADB service
                os.writeBytes(stopCmd + "\n");
                os.flush();
                os.writeBytes(startCmd + "\n");
                os.flush();
                
                os.writeBytes("exit\n");
                os.flush();
                
                process.waitFor();
                
            } catch (Exception e) {
                Log.e(TAG, "Failed to disable root debug mode", e);
            }
        } else {
            // Stop ADB forward service
            stopAdbForwardService();
        }
        
        isDebugEnabled = false;
        updateStatus();
        Toast.makeText(this, "Debug mode disabled", Toast.LENGTH_SHORT).show();
    }
    
    private int getDebugPort() {
        try {
            Process process = Runtime.getRuntime().exec("getprop service.adb.tcp.port");
            BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()));
            String line = reader.readLine();
            if (line != null && !line.isEmpty()) {
                return Integer.parseInt(line.trim());
            }
        } catch (Exception e) {
            Log.e(TAG, "Failed to get debug port", e);
        }
        return 5555; // Default port
    }
    
    private void startAdbForwardService(int port) {
        try {
            // Start ADB forward service, using sni-adb as endpoint name
            Log.i(TAG, "Starting ADB forward service, port: " + port + ", endpoint: sni-adb");
            
            // Here we need to call Go code to start piko client, forward ADB port
            // Due to gomobile limitations, we need to implement through other means
            // In actual implementation, this should call Go's piko client code
            
        } catch (Exception e) {
            Log.e(TAG, "Failed to start ADB forward service", e);
        }
    }
    
    private void stopAdbForwardService() {
        try {
            Log.i(TAG, "Stopping ADB forward service");
            // Stop ADB forward service
        } catch (Exception e) {
            Log.e(TAG, "Failed to stop ADB forward service", e);
        }
    }
    
    private void openDeveloperOptions() {
        Intent intent = new Intent(Settings.ACTION_APPLICATION_DEVELOPMENT_SETTINGS);
        if (intent.resolveActivity(getPackageManager()) != null) {
            startActivity(intent);
        } else {
            Toast.makeText(this, "Unable to open developer options", Toast.LENGTH_SHORT).show();
        }
    }
    
    private void updateStatus() {
        StringBuilder status = new StringBuilder();
        status.append("Status Information:\n");
        status.append("Gottyp Mode: ").append(isGottypRunning ? "Running" : "Stopped").append("\n");
        status.append("Debug Mode: ").append(isDebugEnabled ? "Enabled" : "Disabled").append("\n");
        status.append("Root Status: ").append(hasRoot ? "Granted (" + rootCommand + ")" : "Not Granted").append("\n");
        status.append("Remote Address: ").append(remoteAddrEdit.getText().toString()).append("\n");
        
        // 添加 adb logcat 命令提示
        status.append("\nADB Logcat Commands:\n");
        status.append("• adb logcat -s GottypAndroid\n");
        status.append("• adb logcat | grep GottypAndroid\n");
        status.append("• adb logcat -v time | grep -E '(GottypAndroid|gottyp)'\n");
        
        statusText.setText(status.toString());
    }
    
    /**
     * 执行 adb logcat 命令并输出到日志
     * 这个方法可以在需要时调用，用于调试目的
     */
    private void executeAdbLogcat() {
        try {
            Log.i(TAG, "=== 执行 ADB Logcat 命令 ===");
            
            // 构建 adb logcat 命令
            String[] commands = {
                "adb logcat -s GottypAndroid",
                "adb logcat | grep GottypAndroid", 
                "adb logcat -v time | grep -E '(GottypAndroid|gottyp)'"
            };
            
            for (String cmd : commands) {
                Log.i(TAG, "建议命令: " + cmd);
            }
            
            // 如果设备支持，尝试执行基本的 logcat 命令
            if (hasRoot) {
                Log.i(TAG, "尝试通过 root 权限执行 logcat...");
                Process process = Runtime.getRuntime().exec(rootCommand);
                DataOutputStream os = new DataOutputStream(process.getOutputStream());
                
                os.writeBytes("logcat -d -s GottypAndroid | head -20\n");
                os.flush();
                os.writeBytes("exit\n");
                os.flush();
                
                process.waitFor();
                Log.i(TAG, "Logcat 命令执行完成");
            } else {
                Log.i(TAG, "无 root 权限，无法直接执行 logcat 命令");
                Log.i(TAG, "请使用以下命令在电脑上查看日志:");
                Log.i(TAG, "adb logcat -s GottypAndroid");
            }
            
        } catch (Exception e) {
            Log.e(TAG, "执行 adb logcat 命令时发生异常", e);
        }
    }
    
    @Override
    public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        if (requestCode == PERMISSION_REQUEST_CODE) {
            if (grantResults.length > 0 && grantResults[0] == PackageManager.PERMISSION_GRANTED) {
                Toast.makeText(this, "Permission granted", Toast.LENGTH_SHORT).show();
            } else {
                Toast.makeText(this, "Permission denied", Toast.LENGTH_SHORT).show();
            }
        }
    }
}
