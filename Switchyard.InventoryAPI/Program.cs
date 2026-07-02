using System.Threading.Channels;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.EntityFrameworkCore;
using Microsoft.OpenApi;
using Scalar.AspNetCore;
using System.Text.Json.Nodes;
using Switchyard.InventoryAPI.Data;
using Switchyard.InventoryAPI.Data.Interfaces;
using Switchyard.InventoryAPI.Data.Sync;
using Switchyard.Domain;
using Switchyard.InventoryAPI.Services;
using Switchyard.InventoryAPI.Services.Interfaces;

var builder = WebApplication.CreateBuilder(args);

var httpsPort = builder.Configuration.GetValue<int>("Ports:Https", 7000);
builder.WebHost.UseUrls($"https://localhost:{httpsPort}");

builder.Services.AddCors(options =>
{
    options.AddDefaultPolicy(policy =>
        policy.WithOrigins("https://localhost:5173")
              .AllowAnyHeader()
              .AllowAnyMethod());
});

builder.Services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
    .AddJwtBearer(options =>
    {
        options.Authority = builder.Configuration["Auth0:Authority"];
        options.Audience = builder.Configuration["Auth0:Audience"];
    });

builder.Services.AddAuthorization(options =>
{
    options.AddPolicy("ReadInventory",
        policy => policy.RequireClaim("permissions", "read:inventory"));
});

var writeCs = builder.Configuration.GetConnectionString("InventoryWrite")!;
var readCs  = builder.Configuration.GetConnectionString("InventoryRead")!;

var syncChannel = Channel.CreateUnbounded<SyncJob>();
builder.Services.AddSingleton(syncChannel.Writer);
builder.Services.AddSingleton(syncChannel.Reader);

builder.Services.AddSingleton<InventorySyncInterceptor>();
builder.Services.AddDbContext<InventoryContext>((sp, options) =>
{
    options.UseNpgsql(writeCs);
    options.AddInterceptors(sp.GetRequiredService<InventorySyncInterceptor>());
});
builder.Services.AddDbContext<InventoryReadContext>(options => options.UseNpgsql(readCs));

builder.Services.AddScoped<IUnitOfWork, UnitOfWork>();
builder.Services.AddScoped<IClothingService, ClothingService>();
builder.Services.AddScoped<IPPEService, PPEService>();
builder.Services.AddScoped<IToolService, ToolService>();
builder.Services.AddHostedService<InventorySyncWorker>();
builder.Services.AddControllers();

var auth0Domain    = builder.Configuration["Auth0:Authority"]!.TrimEnd('/');
var auth0Audience  = builder.Configuration["Auth0:Audience"]!;
var scalarClientId = builder.Configuration["Auth0:ScalarClientId"]!;

var logoPath    = Path.Combine(builder.Environment.ContentRootPath, "..", "Switchyard.UI", "src", "assets", "Digital Parts Full Logo Light Mode.svg");
var logoDataUri = File.Exists(logoPath)
    ? $"data:image/svg+xml;base64,{Convert.ToBase64String(File.ReadAllBytes(logoPath))}"
    : string.Empty;

builder.Services.AddOpenApi(options =>
{
    options.AddDocumentTransformer((document, _, _) =>
    {
        document.Components ??= new OpenApiComponents();
        // v2.0: SecuritySchemes expects IDictionary<string, IOpenApiSecurityScheme>
        document.Components.SecuritySchemes = new Dictionary<string, IOpenApiSecurityScheme>
        {
            ["oauth2"] = new OpenApiSecurityScheme
            {
                Type  = SecuritySchemeType.OAuth2,
                Flows = new OpenApiOAuthFlows
                {
                    AuthorizationCode = new OpenApiOAuthFlow
                    {
                        // audience baked in here — Auth0 requires it; not a standard OAuth2 param
                        AuthorizationUrl = new Uri($"{auth0Domain}/authorize?audience={Uri.EscapeDataString(auth0Audience)}"),
                        TokenUrl         = new Uri($"{auth0Domain}/oauth/token"),
                        Scopes           = new Dictionary<string, string>
                        {
                            ["openid"]           = "OpenID Connect",
                            ["profile"]          = "User profile",
                            ["read:inventory"]   = "Read inventory data",
                        },
                    }
                }
            }
        };

        if (!string.IsNullOrEmpty(logoDataUri))
        {
            document.Info ??= new OpenApiInfo();
            document.Info.Extensions ??= new Dictionary<string, IOpenApiExtension>();
            document.Info.Extensions["x-logo"] = new JsonNodeExtension(
                JsonNode.Parse($"{{\"url\":\"{logoDataUri}\",\"altText\":\"Digital Parts\"}}")!);
        }

        // v2.0: Security (not SecurityRequirements); key is OpenApiSecuritySchemeReference
        document.Security =
        [
            new OpenApiSecurityRequirement
            {
                [new OpenApiSecuritySchemeReference("oauth2", document)] = []
            }
        ];

        return Task.CompletedTask;
    });
});

var app = builder.Build();

var logger = app.Services.GetRequiredService<ILogger<Program>>();
const int maxRetries = 10;
var retryDelay = TimeSpan.FromSeconds(3);

for (var attempt = 1; attempt <= maxRetries; attempt++)
{
    try
    {
        using var scope = app.Services.CreateScope();
        var writeCtx = scope.ServiceProvider.GetRequiredService<InventoryContext>();
        var readCtx  = scope.ServiceProvider.GetRequiredService<InventoryReadContext>();
        writeCtx.Database.Migrate();
        readCtx.Database.EnsureCreated();
        break;
    }
    catch (Exception ex) when (attempt < maxRetries)
    {
        logger.LogWarning("DB not ready (attempt {Attempt}/{Max}): {Message}. Retrying in {Delay}s…",
            attempt, maxRetries, ex.Message, retryDelay.TotalSeconds);
        await Task.Delay(retryDelay);
    }
}

// Enqueue an initial full sync so the read DB is populated on startup
syncChannel.Writer.TryWrite(new SyncJob(new HashSet<Type> { typeof(Clothing), typeof(PPE), typeof(Tool) }));

app.UseRouting();
app.UseCors();
app.UseAuthentication();
app.UseAuthorization();
app.MapControllers();
app.MapOpenApi();
app.MapScalarApiReference(options =>
{
    options
        .ForceLightMode()
        .HideDarkModeToggle()
        .AddOAuth2Authentication("oauth2", scheme =>
        {
            scheme.WithFlows(flows =>
                flows.WithAuthorizationCode(flow =>
                    flow.WithClientId(scalarClientId)));
        });
});

app.Run();
