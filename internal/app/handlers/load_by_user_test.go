package handlers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/handlers"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/repository"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/repository/mocks"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

var _ = Describe("LoadByUser", func() {
	var ts *httptest.Server
	var repositoryMock *mocks.URLRepository
	var cfg config.Config

	BeforeEach(func() {
		repositoryMock = new(mocks.URLRepository)
		cfg = config.Config{BaseURL: "http://localhost:8080"}

		service := handlers.NewService(repositoryMock, nil, cfg)
		r := handlers.NewRouter(service)
		ts = httptest.NewServer(r)
	})
	AfterEach(func() {
		ts.Close()
	})

	Context("when userid cookie not specified", func() {
		BeforeEach(func() {
			repositoryMock.On("LoadByUserID", mock.Anything, mock.Anything).Return([]repository.URLEntity{}, nil)
		})
		It("should respond 204", func() {
			res := testGetList(ts, nil)
			Expect(res.StatusCode).To(Equal(204))
		})
		It("should set userid cookie", func() {
			res := testGetList(ts, nil)
			Expect(res.Cookies()).To(HaveLen(1))
			cookie := res.Cookies()[0]
			Expect(cookie.Name).To(Equal("UserID"))
			Expect(cookie.Value).NotTo(BeEmpty())
		})
	})

	Context("when userid cookie provided", func() {
		var cookie *http.Cookie
		var userID string

		BeforeEach(func() {
			// делаем запрос без куки, чтобы получить значение для нее
			repositoryMock.On("LoadByUserID", mock.Anything, mock.Anything).Return([]repository.URLEntity{}, nil).Once()
			res := testGetList(ts, nil)
			cookie = res.Cookies()[0]
			userID = strings.Split(cookie.Value, ":")[0]
		})
		AfterEach(func() {
			cookie = nil
			userID = ""
		})

		When("there are no urls for specified userid", func() {
			BeforeEach(func() {
				repositoryMock.On("LoadByUserID", mock.Anything, userID).Return([]repository.URLEntity{}, nil).Once()
			})

			It("should respond 204 when there are no urls for specified userid", func() {
				res := testGetList(ts, []*http.Cookie{cookie})
				Expect(res.StatusCode).To(Equal(204))
			})
		})

		When("there are urls for specified userid", func() {
			var urlEntities = []repository.URLEntity{
				{
					ID:          "123",
					OriginalURL: "http://google.com",
					UserID:      userID,
				},
				{
					ID:          "456",
					OriginalURL: "http://yandex.ru",
					UserID:      userID,
				},
			}
			BeforeEach(func() {
				repositoryMock.On("LoadByUserID", mock.Anything, userID).Return(urlEntities, nil).Once()
			})

			It("should return urls list for specified user", func() {
				var expectedJson = `[
{
"short_url": "http://localhost:8080/123", 
"original_url": "http://google.com"
}, 
{
"short_url": "http://localhost:8080/456", 
"original_url": "http://yandex.ru"
}
]`
				res := testGetList(ts, []*http.Cookie{cookie})
				Expect(res.StatusCode).To(Equal(200))
				body, err := io.ReadAll(res.Body)
				defer res.Body.Close()
				Expect(err).NotTo(HaveOccurred())
				Expect(body).To(MatchJSON(expectedJson))
			})
		})

	})
})

func testGetList(ts *httptest.Server, cookies []*http.Cookie) *http.Response {
	return testRequest(ts, "GET", "/user/urls", cookies, nil)
}

func testRequest(ts *httptest.Server, method, path string, cookies []*http.Cookie, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, body)

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	Expect(err).NotTo(HaveOccurred())
	c := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := c.Do(req)
	Expect(err).NotTo(HaveOccurred())

	return resp
}
